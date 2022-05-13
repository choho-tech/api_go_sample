package main
import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"strings"
	"encoding/json"
	"bytes"
	"bufio"
)

const (
	BASE_URL = "<your base_url>"
	FILE_SERVER_URL = "<your file_server_url>"
	USER_ID = "<your user_id>"
	ZH_TOKEN = "<your zh_token>"

	FILE_PATH = "l.stl" // 本地stl文件地址
	JAW_TYPE = "Lower" // 上颌为Upper, 下颌为Lower
)


func main() {
	now := time.Now().Unix()

	// Step 1. upload stl to file server
	req, _ := http.NewRequest("GET", FILE_SERVER_URL +
		"/scratch/APIClient/" + USER_ID + "/upload_url?postfix=stl", nil)
	req.Header.Set("X-ZH-TOKEN", ZH_TOKEN)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		panic("get upload url failed " + err.Error())
	}
	defer resp.Body.Close()

	respByte, _ := ioutil.ReadAll(resp.Body)
	respStr := string(respByte)

	file, err := os.Open(FILE_PATH)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	req, err = http.NewRequest(http.MethodPut, respStr[1:len(respStr)-1], file)
	if err != nil {
		panic(err)
	}

	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println("Upload takes", time.Now().Unix()-now, "seconds")

	stl_urn := "urn:zhfile:o:s:APIClient:" + USER_ID + ":" + respStr[
		strings.Index(respStr, USER_ID) + len(USER_ID) + 1:
		strings.Index(respStr, "?")]

	// Step 2. launch job

	type Mesh struct {
		Type 		string	`json:"type"`
		Data		string 	`json:"data"`
	}

	type InputData struct {
		JawType		string	`json:"jaw_type"`
		Mesh		Mesh	`json:"mesh"`
	}

	type OutputConfig struct {
		Type		string `json:"type"`
	}

	type MeshOutConfig struct {
		Mesh		OutputConfig `json:"mesh"`
	}

	type Job struct {
		SpecGroup		string	`json:"spec_group"`
		SpecName		string	`json:"spec_name"`
		SpecVersion		string	`json:"spec_version"`
		UserId 			string	`json:"user_id"`
		UserGroup		string	`json:"user_group"`
		InputData		InputData	`json:"input_data"`
		MeshOutConfig	MeshOutConfig `json:"output_config"`
	}

	job := Job {
		SpecGroup: "mesh-processing",
		SpecName: "oral-seg",
		SpecVersion: "1.0-snapshot",
		UserId:	USER_ID,
		UserGroup:	"APIClient",
		InputData:	InputData {
			JawType:	JAW_TYPE,
			Mesh:		Mesh {
				Type:	"stl",
				Data:	stl_urn,
			},
		},
		MeshOutConfig:	MeshOutConfig {
			Mesh:		OutputConfig {
				Type:	"stl",
			},
		},
	}

	job_json, _ := json.Marshal(job)
	if err != nil {
		panic(err)
	}

	req, err = http.NewRequest("POST", BASE_URL + "/run", bytes.NewBuffer(job_json))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ZH-TOKEN", ZH_TOKEN)

	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respByte, _ = ioutil.ReadAll(resp.Body)

	json_msg := make(map[string]json.RawMessage)
	json.Unmarshal(respByte, &json_msg)
	run_id := string(json_msg["run_id"])
	run_id = run_id[1:len(run_id)-1]
	fmt.Println("run id is", run_id)

	type RunResult struct {
		Completed	bool		`json:"completed"`
		Failed		bool		`json:"failed"`
		Reason		string		`json:"reason_public"`
		X map[string]interface{} `json:"-"`
	}

	now = time.Now().Unix()
	run_json := RunResult{}

	// Step 3. wait until job finished
	for true {
		time.Sleep(time.Duration(3) * time.Second)
		req, _ = http.NewRequest("GET", BASE_URL + "/run/" + run_id, nil)
		req.Header.Set("X-ZH-TOKEN", ZH_TOKEN)
		resp, err = (&http.Client{}).Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		respByte, _ = ioutil.ReadAll(resp.Body)
		json.Unmarshal(respByte, &run_json)

		if(run_json.Failed){
			panic("job failed with error: " + run_json.Reason)
		}

		if(run_json.Completed) {break}
	}

	fmt.Println("Job completed in", time.Now().Unix()-now, "seconds")

	// Step 4. get results
	type DataResult struct {
		SegLab	[]int		`json:"seg_labels"`
		Mesh	Mesh		`json:"mesh"`
		X map[string]interface{} `json:"-"`
	}

	req, _ = http.NewRequest("GET", BASE_URL + "/data/" + run_id, nil)
	req.Header.Set("X-ZH-TOKEN", ZH_TOKEN)
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respByte, _ = ioutil.ReadAll(resp.Body)

	data_json := DataResult{}
	json.Unmarshal(respByte, &data_json)

	file, err = os.OpenFile("seg_labels.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic("failed creating file: " + err.Error())
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range data_json.SegLab {
		_, _ = datawriter.WriteString(fmt.Sprintf("%v", data) + "\n")
	}

	datawriter.Flush()
	file.Close()

	// Step 5. download mesh results from file server
	file, err = os.OpenFile("processed_mesh.stl", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("failed creating file: " + err.Error())
	}

	req, _ = http.NewRequest("GET", FILE_SERVER_URL + "/file/download?urn=" + data_json.Mesh.Data, nil)
	req.Header.Set("X-ZH-TOKEN", ZH_TOKEN)
	resp, err = (&http.Client{}).Do(req)

	if err != nil {
		panic("failed downloading file: " + err.Error())
	}
	defer resp.Body.Close()

	_, _ = io.Copy(file, resp.Body)

	defer file.Close()
	fmt.Println("Completed: Mesh saved to processed_mesh.stl and Label saved to seg_labels.txt")
}
