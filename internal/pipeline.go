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
	Bind     string `mapstructure:"bind"`
	Parallel uint   `mapstructure:"parallel"`
	Command  string `mapstructure:"command"`

	command  *template.Template
	objects  []map[string]string
	commands []string
}

type Pipeline struct {
	Pipeline string                         `mapstructure:"pipeline"`
	Tasks    []Task                         `mapstructure:"tasks"`
	Objects  map[string][]map[string]string `mapstructure:"objects"`

	dir       string
	nameField string
	taskMap   map[string]int
}

func LoadPipeline(fp string) (pl *Pipeline, err error) {
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

	pl = new(Pipeline)
	if err = conf.Unmarshal(&pl); err != nil {
		return nil, err
	}

	pl.taskMap = make(map[string]int, len(pl.Tasks))
	for i := range pl.Tasks {
		name = pl.Tasks[i].Name
		if name == "" {
			return nil, fmt.Errorf("tasks[%d] name is empty", i)
		}

		if !_TaskNameRE.Match([]byte(name)) {
			return nil, fmt.Errorf("tasks[%d] name is invalid: %q", i, name)
		}

		if _, ok = pl.taskMap[name]; !ok {
			pl.taskMap[name] = i
		} else {
			return nil, fmt.Errorf("duplicate task name found: %s", name)
		}
	}
	if pl.dir = conf.GetString("work_dir"); pl.dir == "" {
		pl.dir = DEFAULT_Dir
	}
	pl.dir = filepath.Join(
		pl.dir,
		pl.Pipeline+"_"+time.Now().Format("2006-01-02T15-04-05_")+RandString(8),
	)

	if pl.nameField = conf.GetString("name_field"); pl.nameField == "" {
		pl.nameField = DEFAULT_NameField
	}

	if err = os.MkdirAll(pl.dir, 0755); err != nil {
		return nil, err
	}

	if err = pl.compile(); err != nil {
		return nil, err
	}

	return pl, nil
}

func (pl *Pipeline) compile() (err error) {
	for i := range pl.Tasks {
		if err = pl.compileAt(i); err != nil {
			return fmt.Errorf("compile %s([%d]): %w", pl.Tasks[i].Name, i, err)
		}
	}
	return nil
}

func (pl *Pipeline) compileAt(idx int) (err error) {
	var task *Task

	if idx > len(pl.Tasks)-1 {
		return nil
	}
	task = &pl.Tasks[idx]

	if task.command, err = template.New(task.Name).Parse(task.Command); err != nil {
		return err
	}

	for k := range pl.Objects {
		if k == task.Bind {
			task.objects = pl.Objects[k]
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

func (pl *Pipeline) run(idx int, pn int, objects ...string) (errs []error) {
	var (
		n    int
		list []int
		ch   chan struct{}
		task *Task
	)

	if n = len(pl.Tasks); idx < 0 || idx > n-1 || n == 0 {
		return nil
	}
	task = &pl.Tasks[idx]

	if pn < 0 {
		pn = int(pl.Tasks[idx].Parallel)
	}

	if len(objects) == 0 {
		list = make([]int, 0, len(task.objects))
		for i := range task.objects {
			list = append(list, i)
		}
	} else {
		for i := range task.objects {
			if indexOf(objects, task.objects[i][pl.nameField]) > -1 {
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
				started    bool
				err        error
				objectName string
				bts        []byte
				stdoutFile *os.File
				stderrFile *os.File
			)

			objectName = task.objects[i][pl.nameField]
			now, _ := Jobname(task.Name, objectName)
			// prefix := filepath.Join(pl.dir, name)
			prefix := filepath.Join(pl.dir, task.Name+"__"+objectName)

			defer func() {
				if err != nil {
					errs[i] = fmt.Errorf(
						"%s objects[%d](%s): %w", task.Name, i, objectName, err,
					)
				}

				if started {
					os.Rename(prefix+".logging", prefix+".log")

					if info, e := os.Stat(prefix + ".error"); e == nil {
						if info.Size() == 0 {
							os.Remove(prefix + ".error") // ignore error
						}
					}
				}

				<-ch
				wg.Done()
			}()

			script := prefix + ".sh"
			bts = []byte(DEFAULT_Head + "\n" + task.commands[i] + "\n")
			err = ioutil.WriteFile(script, bts, 0755)
			if err != nil {
				return
			}

			if stdoutFile, err = os.Create(prefix + ".logging"); err != nil {
				return
			}
			defer stdoutFile.Close()
			if stderrFile, err = os.Create(prefix + ".error"); err != nil {
				return
			}
			defer stderrFile.Close()

			stdoutFile.WriteString(
				fmt.Sprintf("#### >>> %s %s\n", now.Format(time.RFC3339), task.Name),
			)

			bts, _ = json.Marshal(task.objects[i])
			stdoutFile.WriteString(fmt.Sprintf("####     %s\n\n", bts))

			cmd := exec.Command("/bin/bash", script)
			cmd.Stdout, cmd.Stderr = stdoutFile, stderrFile
			fmt.Println(">>>", now.Format(time.RFC3339), "start", prefix)

			err, started = cmd.Run(), true
			at := time.Now().Format(time.RFC3339)
			if err == nil {
				fmt.Println("<<<", at, "end", prefix)
				stdoutFile.WriteString(fmt.Sprintf("\n\n#### <<< %s\n", at))
			} else {
				fmt.Printf("<<< %s %s %s: %v", at, "failed", prefix, err)
				stdoutFile.WriteString(fmt.Sprintf("\n#### <<< %s %v\n", at, err))
			}
		}(i)
	}

	wg.Wait()
	return
}

func (pl *Pipeline) RunTask(name string, pn int, objects ...string) (err error) {
	var (
		ok   bool
		idx  int
		errs []error
		strs []string
	)

	if idx, ok = pl.taskMap[name]; !ok {
		return fmt.Errorf("task not found")
	}

	errs = pl.run(idx, pn, objects...)
	strs = make([]string, 0, len(errs))
	for i := range errs {
		if errs[i] != nil {
			strs = append(strs, errs[i].Error())
		}
	}

	if len(strs) > 0 {
		data := map[string]interface{}{
			"taskName": name,
			"number":   len(strs),
			"total":    len(errs),
			"errors":   strs,
		}
		bts, _ := json.Marshal(data)
		err = errors.New(string(bts))
	}

	return err
}
