package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Volume       string `json:"volume"`
	MetaDataPath string `json:"metadatapath"`
	Editor       string `json:"editor"`
	GrepCmd      string `json:"grepcmd"`
}

func NewConfig(filename string) (*Config, error) {

	// 設定ファイルが存在しない場合の処理
	if _, err := os.Stat(filename); err != nil {
		fp, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		defer fp.Close()

		// デフォルト設定
		cfg := &Config{
			Volume:       "./data",
			MetaDataPath: "./.metadata.json",
			Editor:       "vim",
			GrepCmd:      "grep",
		}
		encoder := json.NewEncoder(fp)
		encoder.SetIndent("", "	")
		if err := encoder.Encode(cfg); err != nil {
			return nil, err
		}
	}

	// ファイルの中身を読み取る
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, errors.New("config file is empty")
	}

	// configインスタンスの生成
	cfg := new(Config)
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
