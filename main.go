package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var burpURL = "https://portswigger.net"

func main() {
	resp, err := http.Get("https://portswigger.net/web-security/cross-site-scripting/cheat-sheet")
	if err != nil {
		fmt.Println("ERRORED:", err)
	}
	defer resp.Body.Close()
	body, errBody := ioutil.ReadAll(resp.Body)

	if errBody != nil {
		fmt.Println("BODY ERROR:", errBody)
	}

	bodyString := string(body)
	cheatsheetURL := burpURL + findTheCheatSheetChunk(bodyString)
	loadedJavaScript := loadJavaScriptSheet(cheatsheetURL)
	payloadObjects := extractPayloadsFromJavaScript(loadedJavaScript)
	cleanPayloads := cleanUpPayloads(payloadObjects)
	fmt.Println("Saving to file.")
	saveToFile(cleanPayloads)
}

func findTheCheatSheetChunk(data string) string {
	cheatsheetRE := regexp.MustCompile(`<script.*?src=".*cheat-sheet(.*?)"></script>`)

	// we should only really have ONE finding here, SHOULD only have one anyway.
	cheatsheetURL := cheatsheetRE.FindAllStringSubmatch(data, -1)
	if len(cheatsheetURL) == 0 {
		fmt.Println("Couldn't find the JS cheat sheet, rip")
		os.Exit(1) // didn't find any
	}

	// should be: <script async src="/bundles/public/staticcms/cross-site-scripting/cheat-sheet?v=iCuH1TwKKrx4GksM_6XOxyiRVGGmigVzmiT-BSNo7981"></script>
	return strings.Split(cheatsheetURL[0][0], "\"")[1]
}

func loadJavaScriptSheet(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("ERRORED:", err)
	}
	defer resp.Body.Close()
	body, errBody := ioutil.ReadAll(resp.Body)

	if errBody != nil {
		fmt.Println("BODY ERROR:", errBody)
	}

	bodyString := string(body)
	return bodyString
}

func extractPayloadsFromJavaScript(data string) [][]string {
	dataChunkRE := regexp.MustCompile(`code:(.*?),`)
	foundPayloads := dataChunkRE.FindAllStringSubmatch(data, -1)
	return foundPayloads
}

func cleanUpPayloads(data [][]string) []string {
	allPayloads := []string{}

	// inefficient loops and clean up, nice.
	for _, v := range data {
		payloadStr := string(v[1])
		payloadStr = payloadStr[:len(payloadStr)-1]
		payloadStr = payloadStr[1:]
		payloadStr = strings.Replace(payloadStr, "\\/", "/", -1)
		allPayloads = append(allPayloads, payloadStr)
	}

	return allPayloads
}

func saveToFile(list []string) {
	f, err := os.Create("payloads.txt")
	if err != nil {
		fmt.Println("Could not create file, does it already exist?")
	}

	defer f.Close()

	for _, v := range list {
		_, err := f.WriteString(v + "\n")
		if err != nil {
			fmt.Println("Failed to write to file, wordlist might not be complete.")
		}
	}
	f.Sync()
}
