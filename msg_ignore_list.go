package main

func readMsgIgnoreListFile() ([]string, error) {
	configFile, err := readAppConfigFile()
	if err != nil {
		return nil, err
	}

}

func addToMsgIgnoreList(text string) (string, error) {

}
