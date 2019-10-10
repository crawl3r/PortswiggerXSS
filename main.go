package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"flag"
	"bufio"
)

var burpURL = "https://portswigger.net"

//vars for params
var (
  	tag string
  	event string
	update bool
	filename string
	final []string
)

func main() {
	flag.StringVar(&tag, "tag", "", "tag to filter by")
	flag.StringVar(&event, "event", "", "event to filter by")
	flag.BoolVar(&update, "update", false, "update the payloads file")
	flag.StringVar(&filename, "file", "final.txt", "file name for the filtered payload text file")

	flag.Parse()

	_, err := os.Stat("payloads.txt")
	if (update || os.IsNotExist(err)) {
		downloadPayloadFile()
	}


	filter()
}

func filter() {
	var check = false
	file, err := os.Open("payloads.txt")
	if err != nil {
	    fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
	    line := scanner.Text()
	    if tag != "" {
	    	if event != "" {
	    		if strings.HasPrefix(line, "<" + tag + " ") && strings.Contains(line, " " + event) {
	    			check = true
	    		}
	    	} else {
	    		if strings.HasPrefix(line, "<" + tag + " ") {
	    			check = true
	    		}
	    	}
	    } else {
	    	if event != "" {
	    		if strings.Contains(line, " " + event) {
	    			check = true
	    		}
	   	 	}
	    }
	    
		if check {
			final = append(final, line)
	    	check = false
		}
	}

	saveToFile(final, filename)

	if err := scanner.Err(); err != nil {
	    fmt.Println(err)
	}
	
}

func downloadPayloadFile() {
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
	cheatsheetURL := burpURL + findTheCheatSheet(bodyString)
	loadedJavaScript := loadJavaScriptSheet(cheatsheetURL)
	payloadObjects := extractPayloadsFromJavaScript(loadedJavaScript)
	cleanPayloads := cleanUpPayloads(payloadObjects)
	fmt.Println("Saving to file.")
	saveToFile(cleanPayloads, "payloads.txt")
}



func findTheCheatSheet(data string) string {
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

func saveToFile(list []string, filename string) {
	f, err := os.Create(filename)
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
