package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	// "strings"
	"sync"
	"text/template"
	"time"

	"github.com/spf13/viper"
)

type Task struct {
	Name     string `mapstructure:"name"`
	Kind     string `mapstructure:"kind"`
	Parallel uint   `mapstructure:"parallel"`
	Block    bool   `mapstructure:"block"`
	Command  string `mapstructure:"command"`
	command  *template.Template
	objects  []map[string]string
	commands []string
}

type Pipeline struct {
	Tasks     []Task                         `mapstructure:"tasks"`
	Objects   map[string][]map[string]string `mapstructure:"objects"`
	dir       string
	nameField string
	taskMap   map[string]int
}

func LoadPipeline(fp string) (p *Pipeline, err error) {
	var (
		ok   bool
		name string
		conf *viper.Viper
	)

	conf = viper.New()
	conf.SetConfigName("pipeline")
	conf.SetConfigType("yaml")
	conf.SetConfigFile(fp)

	if err = conf.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	p = new(Pipeline)
	if err = conf.Unmarshal(&p); err != nil {
		return nil, err
	}

	p.taskMap = make(map[string]int, len(p.Tasks))
	for i := range p.Tasks {
		name = p.Tasks[i].Name
		if name == "" {
			return nil, fmt.Errorf("tasks[%d] name is empty", i)
		}

		if !_TaskNameRE.Match([]byte(name)) {
			return nil, fmt.Errorf("tasks[%d] name is invalid: %q", i, name)
		}

		if _, ok = p.taskMap[name]; !ok {
			p.taskMap[name] = i
		} else {
			return nil, fmt.Errorf("duplicate task name found: %s", name)
		}
	}
	if p.dir = conf.GetString("work_dir"); p.dir == "" {
		p.dir = DEFAULT_Dir
	}
	p.dir = filepath.Join(p.dir, time.Now().Format("2006-01-02T15-04-05_")+RandString(8))

	if p.nameField = conf.GetString("name_field"); p.nameField == "" {
		p.nameField = DEFAULT_NameField
	}

	if err = os.MkdirAll(p.dir, 0755); err != nil {
		return nil, err
	}

	if err = p.compile(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pipeline) compile() (err error) {
	for i := range p.Tasks {
		if err = p.compileAt(i); err != nil {
			return fmt.Errorf("compile %s([%d]): %w", p.Tasks[i].Name, i, err)
		}
	}
	return nil
}

func (p *Pipeline) compileAt(idx int) (err error) {
	var task *Task

	if idx > len(p.Tasks)-1 {
		return nil
	}
	task = &p.Tasks[idx]

	if task.command, err = template.New(task.Name).Parse(task.Command); err != nil {
		return err
	}

	for k := range p.Objects {
		if k == task.Kind {
			task.objects = p.Objects[k]
		}
	}

	n := len(task.objects)
	task.commands = make([]string, n)

	buf := new(bytes.Buffer)
	for i := range task.objects {
		if err = task.command.Execute(buf, task.objects[i]); err != nil {
			return err
		}
		task.commands[i] = buf.String()
		buf.Reset()
	}

	return
}

func (p *Pipeline) run(idx int, pn int, objects ...string) (errs []error) {
	var (
		n    int
		list []int
		ch   chan struct{}
		task *Task
	)

	if n = len(p.Tasks); idx < 0 || idx > n-1 || n == 0 {
		return nil
	}
	task = &p.Tasks[idx]

	if pn < 0 {
		pn = int(p.Tasks[idx].Parallel)
	}

	if len(objects) == 0 {
		list = make([]int, 0, len(task.objects))
		for i := range task.objects {
			list = append(list, i)
		}
	} else {
		for i := range task.objects {
			if indexOf(objects, task.objects[i][p.nameField]) > -1 {
				list = append(list, i)
			}
		}
	}

	n = len(list)
	errs = make([]error, n)
	wg := new(sync.WaitGroup)

	if pn == 0 {
		ch = make(chan struct{}, n)
	} else {
		ch = make(chan struct{}, pn)
	}

	for _, i := range list {
		ch <- struct{}{}
		wg.Add(1)
		go func(i int) {
			var (
				err        error
				objectName string
				logFile    *os.File
			)

			objectName = task.objects[i][p.nameField]

			defer func() {
				if err != nil {
					errs[i] = fmt.Errorf(
						"%s objects[%d](%s): %w", task.Name, i, objectName, err,
					)
				}
				<-ch
				wg.Done()
			}()

			now, name := Jobname(task.Name, objectName)
			prefix := filepath.Join(p.dir, name)
			script := prefix + ".sh"
			err = ioutil.WriteFile(script, []byte(DEFAULT_Head+task.commands[i]+"\n"), 0755)
			if err != nil {
				return
			}

			if logFile, err = os.Create(prefix + ".log"); err != nil {
				return
			}
			defer logFile.Close()
			logFile.WriteString(
				fmt.Sprintf("#### >>> %s %s\n", now.Format(time.RFC3339), task.Name),
			)

			bts, _ := json.Marshal(task.objects[i])
			logFile.WriteString(fmt.Sprintf("####     %s\n\n", bts))

			cmd := exec.Command("/bin/bash", script)
			cmd.Stdout, cmd.Stderr = logFile, logFile
			fmt.Println(">>>", now.Format(time.RFC3339), "start", prefix)

			err = cmd.Run()
			at := time.Now().Format(time.RFC3339)
			if err == nil {
				fmt.Println("<<<", at, "end", prefix)
				logFile.WriteString(fmt.Sprintf("\n\n#### <<< %s\n", at))
			} else {
				fmt.Printf("<<< %s %s %s: %v", at, "failed", prefix, err)
				logFile.WriteString(fmt.Sprintf("\n#### <<< %s %v\n", at, err))
			}
		}(i)
	}

	wg.Wait()
	return
}

func (p *Pipeline) RunTask(name string, pn int, objects ...string) (err error) {
	var (
		ok   bool
		idx  int
		errs []error
		strs []string
	)

	if idx, ok = p.taskMap[name]; !ok {
		return fmt.Errorf("task not found")
	}

	errs = p.run(idx, pn, objects...)
	strs = make([]string, 0, len(errs))
	for i := range errs {
		if errs[i] != nil {
			strs = append(strs, errs[i].Error())
		}
	}

	if len(strs) > 0 {
		data := map[string]interface{}{
			"title":    fmt.Sprintf("%d task(s) failed", len(strs)),
			"taskName": name,
			"number":   len(strs),
			"errors":   strs,
		}
		bts, _ := json.Marshal(data)
		err = errors.New(string(bts))
	}

	return err
}
