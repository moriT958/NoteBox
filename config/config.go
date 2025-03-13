package config

import (
	"encoding/json"
	"errors"
	"os"
)

type config struct {
	Volume       string `json:"volume"`
	MetaDataPath string `json:"metadatapath"`
	Editor       string `json:"editor"`
	GrepCmd      string `json:"grepcmd"`
}

func LoadConfigFile(filename string) error {

	// 設定ファイルが存在しない場合の処理
	if _, err := os.Stat(filename); err != nil {
		fp, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer fp.Close()

		cfg := &config{
			Volume:       "",
			MetaDataPath: "",
			Editor:       "",
			GrepCmd:      "",
		}
		encoder := json.NewEncoder(fp)
		encoder.SetIndent("", "	")
		if err := encoder.Encode(cfg); err != nil {
			return err
		}
	}

	// ファイルの中身を読み取る
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return errors.New("config file is empty")
	}

	// configインスタンスの生成
	cfg := new(config)
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return err
	}

	return nil
}
