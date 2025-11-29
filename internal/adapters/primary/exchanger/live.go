package exchanger

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"

	"marketflow/pkg/conc"
)

type LiveExchanger struct {
	Name          string
	Host          string
	Port          string
	receivedTasks int
	wg            *sync.WaitGroup
	cancel        context.CancelFunc
}

func NewLiveExchanger(name, host, port string) (*LiveExchanger, error) {
	if host == "" || port == "" {
		return nil, nil
	}

	return &LiveExchanger{
		Name:          name,
		Host:          host,
		Port:          port,
		receivedTasks: 0,
		wg:            &sync.WaitGroup{},
	}, nil
}

func (l *LiveExchanger) Stream(ctx context.Context, out chan<- conc.Task, results chan<- Result) {
	ctx, cancel := context.WithCancel(ctx)

	l.cancel = cancel
	defer cancel()

	select {
	case <-ctx.Done():
		l.sendResult(results, ctx.Err())
		return
	default:
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(l.Host, l.Port))
	if err != nil {
		l.sendResult(results, err)
		return
	}

	if err = l.handle(ctx, conn, out); err != nil {
		l.sendResult(results, err)
		return
	}
	l.sendResult(results, nil)
}

func (l *LiveExchanger) Stop() error {
	if l.cancel != nil {
		l.cancel()
		return nil
	}

	return fmt.Errorf("exchanger %s not running", l.Name)
}

func (l *LiveExchanger) handle(ctx context.Context, conn net.Conn, out chan<- conc.Task) error {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		case out <- conc.WrapTask(l.Name, scanner.Text()):
			l.receivedTasks++
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return fmt.Errorf("connection to exchanger %s closed", l.Name)
}

func (l *LiveExchanger) sendResult(results chan<- Result, err error) {
	results <- Result{Name: l.Name, Host: l.Host, Port: l.Port, ReceivedTasks: l.receivedTasks, Err: err}
}
