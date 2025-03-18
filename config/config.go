package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

func LoadConfigFile() error {

	// 設定ファイルが存在しない場合の処理
	if _, err := os.Stat(ConfigFile); err != nil {
		if err := loadDefaultConfig(); err != nil {
			return err
		}
	}

	fp, err := os.Open(ConfigFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	var tempConfig map[string]string
	if err := json.NewDecoder(fp).Decode(&tempConfig); err != nil {
		return err
	}

	if v, ok := tempConfig["volume"]; ok {
		Volume = v
	}
	if v, ok := tempConfig["metadatadir"]; ok {
		MetadataDir = v
	}
	if v, ok := tempConfig["editor"]; ok {
		Editor = v
	}
	if v, ok := tempConfig["grepcmd"]; ok {
		Grepcmd = v
	}

	return nil
}

func loadDefaultConfig() error {
	// デフォルトの設定
	defaultConfig := map[string]string{
		"volume":      "./.data",
		"metadatadir": "./db.sqlite",
		"editor":      "vim",
		"grepcmd":     "grep",
	}

	fp, err := os.Create(ConfigFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	encoder := json.NewEncoder(fp)
	encoder.SetIndent("", "    ") // JSON を整形
	if err := encoder.Encode(defaultConfig); err != nil {
		return err
	}
	fmt.Printf("Config file not found, created default config: %s", ConfigFile)

	return nil
}
