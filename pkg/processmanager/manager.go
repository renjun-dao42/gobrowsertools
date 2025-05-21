package processmanager

import (
	"agent/utils/log"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Process 表示一个进程及其元数据
type Process struct {
	Name           string            // 进程名称
	Cmd            *exec.Cmd         // 当前运行的命令
	Args           []string          // 命令参数
	MaxRestarts    int               // 最大重启次数
	Restarts       int               // 当前重启次数
	Executable     string            // 可执行文件路径
	DependsOn      string            // 依赖的进程名称
	RestartTimeout time.Duration     // 重启超时时间
	stopChan       chan struct{}     // 通知监控协程停止
	readers        []*PrefixedReader // 跟踪所有的读取器
	mutex          sync.Mutex        // 保护Process中的可变字段
}

// 定义一个进程管理器
type ProcessManager struct {
	processes []*Process
	wg        sync.WaitGroup
	mutex     sync.Mutex
	stopping  bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// 创建新的进程管理器
func NewProcessManager() *ProcessManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProcessManager{
		processes: make([]*Process, 0),
		stopping:  false,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// 添加进程
func (pm *ProcessManager) AddProcess(name, executable string, args []string, dependsOn string, maxRestarts int) *Process {
	process := &Process{
		Name:           name,
		Executable:     executable,
		Args:           args,
		MaxRestarts:    maxRestarts,
		DependsOn:      dependsOn,
		RestartTimeout: 5 * time.Second,
		stopChan:       make(chan struct{}),
		readers:        make([]*PrefixedReader, 0),
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.processes = append(pm.processes, process)
	return process
}

// 设置重启超时时间
func (p *Process) SetRestartTimeout(timeout time.Duration) {
	p.RestartTimeout = timeout
}

// 创建前缀输出读取器
type PrefixedReader struct {
	prefix string
	reader io.Reader
	writer io.Writer
	done   chan struct{}
}

// 创建新的前缀输出读取器
func NewPrefixedReader(prefix string, reader io.Reader, writer io.Writer) *PrefixedReader {
	prefix2 := fmt.Sprintf("playground_id=%s,docker_id=%s,process_name=%s", os.Getenv("paas_playground_id"), os.Getenv("paas_docker_id"), prefix)
	return &PrefixedReader{
		prefix: prefix2,
		reader: reader,
		writer: writer,
		done:   make(chan struct{}),
	}
}

// 启动前缀输出读取
func (pr *PrefixedReader) Start() {
	scanner := bufio.NewScanner(pr.reader)

	// 增加缓冲区大小，避免长行导致扫描失败
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	go func() {
		defer close(pr.done)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(pr.writer, "[%s] %s\n", pr.prefix, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(pr.writer, "[%s] 读取输出错误: %v\n", pr.prefix, err)
		}
	}()
}

// 等待读取完成
func (pr *PrefixedReader) Wait() {
	<-pr.done
}

// 创建进程的命令
func (pm *ProcessManager) createCommand(p *Process) *exec.Cmd {
	// 使用主上下文来创建命令，确保能正确取消
	cmd := exec.CommandContext(pm.ctx, p.Executable, p.Args...)

	if p.Name == "Rfbproxy" {
		cmd.Env = append(os.Environ(), "RUST_LOG=info")
	}

	// 创建管道来捕获标准输出和标准错误
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Infof("警告: 无法创建标准输出管道: %v\n", err)
		cmd.Stdout = os.Stdout
	} else {
		// 创建带前缀的输出读取器
		stdoutReader := NewPrefixedReader(p.Name, stdoutPipe, os.Stdout)
		stdoutReader.Start()
		p.mutex.Lock()
		p.readers = append(p.readers, stdoutReader)
		p.mutex.Unlock()
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Infof("警告: 无法创建标准错误管道: %v\n", err)
		cmd.Stderr = os.Stderr
	} else {
		// 创建带前缀的错误读取器
		stderrReader := NewPrefixedReader(p.Name+"-错误", stderrPipe, os.Stderr)
		stderrReader.Start()
		p.mutex.Lock()
		p.readers = append(p.readers, stderrReader)
		p.mutex.Unlock()
	}

	p.Cmd = cmd
	return cmd
}

// 启动所有进程
func (pm *ProcessManager) StartAll() error {
	// 设置信号处理器已在main函数中设置

	// 按依赖顺序找到并启动DBUS进程
	var dbusProcess *Process
	for _, p := range pm.processes {
		if p.DependsOn == "" {
			dbusProcess = p
			break
		}
	}

	if dbusProcess == nil {
		return fmt.Errorf("找不到基础进程（没有依赖的进程）")
	}

	// 启动DBUS进程
	log.Infof("启动 %s...\n", dbusProcess.Name)
	cmd := pm.createCommand(dbusProcess)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 %s 失败: %w", dbusProcess.Name, err)
	}
	log.Infof("%s 已启动，PID: %d\n", dbusProcess.Name, cmd.Process.Pid)

	// 给DBUS进程一点时间来初始化
	time.Sleep(500 * time.Millisecond)

	// 监控DBUS进程，也添加到WaitGroup中
	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		pm.monitorProcess(dbusProcess)
	}()

	// 启动其他依赖进程
	for _, p := range pm.processes {
		if p != dbusProcess {
			pm.wg.Add(1)
			go func(p *Process) {
				defer pm.wg.Done()
				pm.startAndMonitor(p)
			}(p)
		}
	}

	return nil
}

// 启动并监控一个进程
func (pm *ProcessManager) startAndMonitor(p *Process) {
	// 检查依赖进程是否存在
	if p.DependsOn != "" {
		// 查找依赖进程
		var dependsOnProcess *Process
		for _, proc := range pm.processes {
			if proc.Name == p.DependsOn {
				dependsOnProcess = proc
				break
			}
		}

		if dependsOnProcess == nil {
			log.Infof("错误: %s 依赖的进程 %s 不存在，无法启动\n", p.Name, p.DependsOn)
			return
		}

		// 确保依赖进程有一个正在运行的命令
		if dependsOnProcess.Cmd == nil || dependsOnProcess.Cmd.Process == nil {
			log.Infof("错误: %s 依赖的进程 %s 未在运行，无法启动\n", p.Name, p.DependsOn)
			return
		}

		// 尝试检查进程是否仍在运行
		if err := dependsOnProcess.Cmd.Process.Signal(syscall.Signal(0)); err != nil {
			log.Infof("错误: %s 依赖的进程 %s 未在运行，无法启动\n", p.Name, p.DependsOn)
			return
		}
	}

	log.Infof("启动 %s...\n", p.Name)
	cmd := pm.createCommand(p)
	err := cmd.Start()
	if err != nil {
		log.Infof("启动 %s 失败: %v\n", p.Name, err)
		return
	}
	log.Infof("%s 已启动，PID: %d\n", p.Name, cmd.Process.Pid)

	pm.monitorProcess(p)
}

// 监控进程并在需要时重启
func (pm *ProcessManager) monitorProcess(p *Process) {
	for {
		// 等待进程结束或者收到停止信号
		processDone := make(chan error, 1)
		waitDone := make(chan struct{})

		go func() {
			processDone <- p.Cmd.Wait()
			close(waitDone)
		}()

		select {
		case <-p.stopChan:
			// 收到停止信号
			// 等待Wait goroutine结束避免泄漏
			<-waitDone
			return
		case err := <-processDone:
			// 进程已结束
			pm.mutex.Lock()

			// 如果管理器正在停止，不重启
			if pm.stopping {
				pm.mutex.Unlock()
				return
			}

			if err != nil {
				log.Infof("%s 异常退出: %v\n", p.Name, err)

				// 检查是否达到最大重启次数
				p.mutex.Lock()
				reachedMaxRestarts := p.Restarts >= p.MaxRestarts && p.MaxRestarts > 0
				p.mutex.Unlock()

				if reachedMaxRestarts {
					log.Infof("%s 已达到最大重启次数 %d，不再重启\n", p.Name, p.MaxRestarts)

					// 如果是基础进程且不重启，终止依赖它的进程
					if p.DependsOn == "" {
						log.Infof("基础进程 %s 终止，将停止所有依赖它的进程\n", p.Name)
						for _, dep := range pm.processes {
							if dep.DependsOn == p.Name && dep.Cmd != nil && dep.Cmd.Process != nil {
								log.Infof("终止依赖进程 %s (PID: %d)...\n", dep.Name, dep.Cmd.Process.Pid)
								// 发送停止信号，但需要先检查通道是否已关闭
								select {
								case <-dep.stopChan:
									// 通道已关闭，不需要再次关闭
								default:
									close(dep.stopChan)
								}
								// 给进程发送SIGTERM信号
								if err := dep.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
									log.Infof("无法发送SIGTERM到 %s: %v\n", dep.Name, err)
								}
							}
						}
					}

					pm.mutex.Unlock()
					return
				}

				// 重启进程
				p.mutex.Lock()
				p.Restarts++
				restartCount := p.Restarts // 在解锁前保存计数
				restartTimeout := p.RestartTimeout
				p.mutex.Unlock()

				log.Infof("正在重启 %s (第 %d 次尝试)...\n", p.Name, restartCount)

				// 在重启前释放锁，避免长时间持有锁
				pm.mutex.Unlock()

				// 等待一小段时间再重启，避免太快重启可能导致的问题
				select {
				case <-time.After(restartTimeout):
				case <-p.stopChan:
					return
				}

				// 清理旧的readers资源
				p.mutex.Lock()
				oldReaders := p.readers
				p.readers = make([]*PrefixedReader, 0)
				p.mutex.Unlock()

				// 创建新的命令
				cmd := pm.createCommand(p)

				startErr := cmd.Start()
				if startErr != nil {
					log.Infof("重启 %s 失败: %v\n", p.Name, startErr)

					// 等待一段时间后再次尝试
					select {
					case <-time.After(3 * time.Second):
						continue
					case <-p.stopChan:
						return
					}
				}

				log.Infof("%s 已重启，新 PID: %d\n", p.Name, cmd.Process.Pid)

				// 等待旧的readers完成工作
				for _, reader := range oldReaders {
					reader.Wait()
				}
			} else {
				log.Infof("%s 已正常退出\n", p.Name)

				// 如果是基础进程且正常退出，终止依赖它的进程
				if p.DependsOn == "" {
					log.Infof("基础进程 %s 正常退出，将停止所有依赖它的进程\n", p.Name)
					for _, dep := range pm.processes {
						if dep.DependsOn == p.Name && dep.Cmd != nil && dep.Cmd.Process != nil {
							log.Infof("终止依赖进程 %s (PID: %d)...\n", dep.Name, dep.Cmd.Process.Pid)
							// 发送停止信号，但需要先检查通道是否已关闭
							select {
							case <-dep.stopChan:
								// 通道已关闭，不需要再次关闭
							default:
								close(dep.stopChan)
							}
							// 给进程发送SIGTERM信号
							if err := dep.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
								log.Infof("无法发送SIGTERM到 %s: %v\n", dep.Name, err)
							}
						}
					}
				}

				pm.mutex.Unlock()
				return
			}
		case <-pm.ctx.Done():
			// 父上下文被取消
			// 等待Wait goroutine结束避免泄漏
			<-waitDone
			return
		}
	}
}

// 等待所有进程结束
func (pm *ProcessManager) Wait() {
	pm.wg.Wait()
}

// 停止所有进程
func (pm *ProcessManager) StopAll() {
	// 首先设置停止标志并取消上下文
	pm.mutex.Lock()
	if pm.stopping {
		// 避免重复调用StopAll
		pm.mutex.Unlock()
		return
	}
	pm.stopping = true
	pm.mutex.Unlock()

	log.Println("正在停止所有进程...")

	// 取消上下文，通知所有进程停止
	pm.cancel()

	// 通知所有监控协程停止
	pm.mutex.Lock()
	for _, p := range pm.processes {
		// 在锁保护下检查通道是否已关闭
		select {
		case <-p.stopChan:
			// 通道已关闭，不需要再次关闭
		default:
			close(p.stopChan)
		}
	}
	pm.mutex.Unlock()

	// 给进程一点时间来优雅退出
	gracePeriod := 5 * time.Second
	gracefulShutdown := make(chan struct{})
	go func() {
		pm.wg.Wait()
		close(gracefulShutdown)
	}()

	// 等待进程优雅退出或者超时
	select {
	case <-gracefulShutdown:
		log.Println("所有进程已优雅终止")
	case <-time.After(gracePeriod):
		log.Println("优雅终止超时，发送SIGTERM信号")

		// 发送SIGTERM信号给所有仍在运行的进程
		// 首先停止依赖其他进程的进程，然后停止基础进程
		dependentProcesses := make([]*Process, 0)
		baseProcesses := make([]*Process, 0)

		pm.mutex.Lock()
		for _, p := range pm.processes {
			if p.DependsOn != "" {
				dependentProcesses = append(dependentProcesses, p)
			} else {
				baseProcesses = append(baseProcesses, p)
			}
		}
		pm.mutex.Unlock()

		// 先停止依赖进程
		for _, p := range dependentProcesses {
			if p.Cmd != nil && p.Cmd.Process != nil {
				log.Infof("正在终止 %s (PID: %d)...\n", p.Name, p.Cmd.Process.Pid)
				if err := p.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
					log.Infof("无法发送 SIGTERM 到 %s: %v\n", p.Name, err)
				}
			}
		}

		// 等待一小段时间
		time.Sleep(1 * time.Second)

		// 再停止基础进程
		for _, p := range baseProcesses {
			if p.Cmd != nil && p.Cmd.Process != nil {
				log.Infof("正在终止 %s (PID: %d)...\n", p.Name, p.Cmd.Process.Pid)
				if err := p.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
					log.Infof("无法发送 SIGTERM 到 %s: %v\n", p.Name, err)
				}
			}
		}

		// 再给进程一点时间来响应SIGTERM
		select {
		case <-gracefulShutdown:
			log.Println("所有进程已响应SIGTERM并终止")
		case <-time.After(2 * time.Second):
			log.Println("部分进程未响应SIGTERM，强制终止")
			for _, p := range pm.processes {
				if p.Cmd != nil && p.Cmd.Process != nil {
					// 查看进程是否还在运行
					if err := p.Cmd.Process.Signal(syscall.Signal(0)); err == nil {
						// 进程仍在运行，强制终止
						log.Infof("强制终止 %s (PID: %d)...\n", p.Name, p.Cmd.Process.Pid)
						p.Cmd.Process.Kill()
					}
				}
			}
		}
	}

	// 等待所有读取器完成
	for _, p := range pm.processes {
		p.mutex.Lock()
		readers := p.readers
		p.mutex.Unlock()
		for _, reader := range readers {
			reader.Wait()
		}
	}

	log.Println("所有进程已终止")
}
