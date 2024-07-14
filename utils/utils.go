package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"pond/pond/globals"
	"pond/pond/templates"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

func RunB(logger zerolog.Logger, command []string, logfile string) error {
	_, err := run(logger, command, "", true, logfile)
	return err
}

func RunI(logger zerolog.Logger, command []string, input string) error {
	_, err := run(logger, command, input, false, "")
	return err
}

func Run(logger zerolog.Logger, command []string) error {
	_, err := run(logger, command, "", false, "")
	return err
}

func RunO(logger zerolog.Logger, command []string) ([]byte, error) {
	return run(logger, command, "", false, "")
}

func run(
	logger zerolog.Logger,
	command []string,
	input string,
	background bool,
	logfile string,
) ([]byte, error) {
	if command[0] == "docker" && command[1] == "container" && command[2] == "create" {
		user, err := user.Current()
		if err != nil {
			logger.Err(err).Msg("")
			return nil, err
		}

		command = append(
			[]string{"docker", "container", "create", "-e", "USER=" + user.Uid}, command[3:]...,
		)
	}

	logger.Trace().
		Str("command", (strings.Join(command, " "))).
		Msg("run command")

	var stderr, stdout bytes.Buffer

	cmd := exec.Command(command[0], command[1:]...)
	if background {
		file, err := os.Create(logfile)
		if err != nil {
			return nil, err
		}
		cmd.Stderr = file
		cmd.Stdout = file
	} else {
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
	}

	if input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			logger.Err(err).Msg("")
			return nil, err
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, input)
		}()
	}

	if background {
		err := cmd.Start()
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	err := cmd.Run()
	if err != nil {
		logger.Err(err).Msg(stderr.String())
		return stderr.Bytes(), err
	}

	return stdout.Bytes(), nil
}

func CopyFile(logger zerolog.Logger, src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}

	err = out.Sync()
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}

	return nil
}

func CopyDir(logger zerolog.Logger, src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	stat, err := os.Stat(src)
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}
	if !stat.IsDir() {
		err := fmt.Errorf("source is not a directory")
		logger.Err(err).Msg("")
		return err
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	if err == nil {
		err = fmt.Errorf("destination already exists")
		logger.Err(err).Msg("")
		return err
	}

	err = os.MkdirAll(dst, stat.Mode())
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(logger, srcPath, dstPath)
			if err != nil {
				logger.Err(err).Msg("")
				return err
			}
		} else {
			err = CopyFile(logger, srcPath, dstPath)
			if err != nil {
				logger.Err(err).Msg("")
				return err
			}
		}
	}

	return nil
}

func GetVersion(logger zerolog.Logger, app string) (string, error) {
	version, found := globals.Versions[app]
	if !found {
		err := fmt.Errorf("version not found")
		logger.Err(err).
			Str("type", app).
			Msg("")
		return "", err
	}

	return version, nil
}

func JsonMerge(data1, data2 []byte) ([]byte, error) {
	var iface1, iface2 interface{}

	err := json.Unmarshal(data1, &iface1)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data2, &iface2)
	if err != nil {
		return nil, err
	}

	m, err := merge(iface1, iface2)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func merge(data1, data2 interface{}) (interface{}, error) {
	map1, ok1 := data1.(map[string]interface{})
	map2, ok2 := data2.(map[string]interface{})

	if !ok1 && !ok2 {
		return data2, nil
	}

	keys := map[string]struct{}{}
	for key := range map1 {
		keys[key] = struct{}{}
	}

	for key := range map2 {
		_, found := keys[key]
		if !found {
			map1[key] = map2[key]
			continue
		}

		value, err := merge(map1[key], map2[key])
		if err != nil {
			fmt.Println(err)
		}

		map1[key] = value
	}

	return map1, nil
}

func Template(src, dst string, data any) error {
	content, err := templates.Templates.ReadFile(src)
	if err != nil {
		return err
	}

	tmpl, err := template.New(dst).Parse(string(content))
	if err != nil {
		return err
	}

	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func HttpGet(logger zerolog.Logger, url string) ([]byte, error) {
	sleep := time.Millisecond * 0

	remaining := time.Second * 10

	retry := true

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Err(err).Msg("")
	}

	var resp *http.Response

	for retry {
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			logger.Err(err).Msg("")
			return nil, err
		}

		if resp.StatusCode == 429 {
			if sleep == 0 {
				sleep = time.Millisecond * 125
			} else {
				sleep = sleep * 2
			}

			remaining = remaining - sleep
			if remaining < 0 {
				return nil, fmt.Errorf("retry timeout reached")
			}

			logger.Info().Dur("sleep", sleep).Msg("hit rate limit")

			time.Sleep(sleep)
			continue
		}

		retry = false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("")
		return nil, err
	}

	resp.Body.Close()

	return body, nil
}

func Sha256(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return fmt.Sprintf("%X", hash.Sum(nil))
}

func ListDiff(s1, s2 []string) []string {
	missing := []string{}

	temp := make(map[string]struct{}, len(s1))
	for _, s := range s1 {
		temp[s] = struct{}{}
	}

	for _, s := range s2 {
		_, found := temp[s]
		if !found {
			missing = append(missing, s)
		}
	}

	return missing
}

func CheckTxResponse(data []byte) (string, error) {
	type Tx struct {
		Code   int    `yaml:"code"`
		Hash   string `yaml:"txhash"`
		RawLog string `yaml:"raw_log"`
	}

	var tx Tx

	err := yaml.Unmarshal(data, &tx)
	if err != nil {
		return tx.Hash, err
	}

	if tx.Code != 0 {
		return tx.Hash, fmt.Errorf(tx.RawLog)
	}

	return tx.Hash, nil
}

func NewTxMsg(msgs []byte) ([]byte, error) {
	msg := fmt.Sprintf(`
	{
		"body": {
			"messages": %s,
			"memo": "",
			"timeout_height": "0",
			"extension_options": [],
			"non_critical_extension_options": []
		},
		"auth_info": {
			"signer_infos": [],
			"fee": {
			"amount": [],
			"gas_limit": "100000000",
			"payer": "",
			"granter": ""
			},
			"tip": null
		},
		"signatures": []
		}
	`, string(msgs))

	return []byte(msg), nil
}
