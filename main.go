package main

import (
	"os"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"bytes"
)

func walk(dirpath string, ifSymLink bool) {
	installed := make(map[string]bool)
	filepath.Walk(dirpath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return nil
		}

		if f.IsDir() {
			return nil
		}

		if f.Mode()&os.ModeSymlink != 0 {
			symlinkPath, err := filepath.EvalSymlinks(f.Name())
			if err != nil {
				println("Error:", err.Error())
			}
			walk(symlinkPath, true)
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			if ifSymLink {
				path = strings.TrimPrefix(path, dirpath)
				path = filepath.Join(filepath.Base(dirpath), path)
			}
			dir := filepath.Dir(path)
			if installed[dir] {
				return nil
			}
			installed[dir] = true
			println("Process: go install", dir)
			cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && go install", dir))
			cmd.Stdout = os.Stdout
			errBuf := bytes.NewBufferString("")
			cmd.Stderr = errBuf
			err := cmd.Run()
			if err != nil {
				println("Error:", err.Error())
			}

			errinfo := errBuf.String()
//			println("Error info:", errinfo)
			if errinfo == "" {
				return nil
			}

			// remove unused imports
			lines := strings.Split(errinfo, "\n")
			for i, _ := range lines {
				line := lines[len(lines) - i - 1]
				if !strings.Contains(line, "imported and not used") {
					continue
				}

				elems := strings.Split(line, ":")
				filename := elems[0]
				lineno := elems[1]

				c := fmt.Sprintf("cd %s && sed -i '%sd' %s", dir, lineno, filename)
				println("Run:", c)
				cmd := exec.Command("sh", "-c", c)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					println("Error:", err.Error())
				}
			}
		}
		return nil
	})
}

func main() {
	walk(".", false)
}
