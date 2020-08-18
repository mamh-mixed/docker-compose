package ecs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/cli"
)

func (b *ecsAPIService) Logs(ctx context.Context, options *cli.ProjectOptions, writer io.Writer) error {
	name := options.Name
	if name == "" {
		project, err := cli.ProjectFromOptions(options)
		if err != nil {
			return err
		}
		name = project.Name
	}

	consumer := logConsumer{
		colors: map[string]ColorFunc{},
		width:  0,
		writer: writer,
	}
	err := b.SDK.GetLogs(ctx, name, consumer.Log)
	if err != nil {
		return err
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	return nil
}

func (l *logConsumer) Log(service, container, message string) {
	cf, ok := l.colors[service]
	if !ok {
		cf = <-Rainbow
		l.colors[service] = cf
		l.computeWidth()
	}
	prefix := fmt.Sprintf("%-"+strconv.Itoa(l.width)+"s |", service)

	for _, line := range strings.Split(message, "\n") {
		buf := bytes.NewBufferString(fmt.Sprintf("%s %s\n", cf(prefix), line))
		l.writer.Write(buf.Bytes())
	}
}

func (l *logConsumer) computeWidth() {
	width := 0
	for n := range l.colors {
		if len(n) > width {
			width = len(n)
		}
	}
	l.width = width + 3
}

type logConsumer struct {
	colors map[string]ColorFunc
	width  int
	writer io.Writer
}
