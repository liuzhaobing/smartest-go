package test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"os"
	asrcloudminds "smartest-go/proto/asr"
	asrcontrol "smartest-go/proto/asrctl"
	"smartest-go/proto/common"
	"strings"
	"testing"
	"time"
)

func asrCloudMindsCallOnce(addr, agentID, path string) string {
	blockSize := 1024
	traceID := uuid.New().String()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(traceID, "[ERROR] AsrCloudMinds Client grpc.Dial error: ", err)
		return err.Error()
	}
	defer conn.Close()
	c := asrcloudminds.NewAsrServiceClient(conn)
	stream, err := c.AsrDecoder(context.Background())
	if err != nil {
		fmt.Println(traceID, "[ERROR] AsrCloudMinds Client AsrDecoder error: ", err)
		return err.Error()
	}

	wavData, _ := ioutil.ReadFile(path)
	if strings.Contains(path, ".wav") {
		wavData = wavData[44:len(wavData)]
	}
	requestBody := &asrcloudminds.AsrDecoderRequest{
		WavData: make([]byte, blockSize),
		WavSize: int32(blockSize),
		WavEnd:  int32(0),
		Guid:    traceID,
		AgentId: agentID,
		RobotId: traceID,
	}
	// 发数据流
	go func() {
		for i := 0; i < len(wavData)+1; i += blockSize {
			v := cap(wavData)
			if i+blockSize <= v {
				requestBody.WavData = wavData[i : i+blockSize]
			} else {
				requestBody.WavData = wavData[i:v]
			}
			err = stream.Send(requestBody)
			if err != nil {
				fmt.Println(traceID, "[ERROR] AsrCloudMinds Client stream.Send error: ", err)
				break
			}
		}
		var emptyData []byte
		requestBody.WavData = emptyData
		err = stream.Send(requestBody)

		if err != nil {
			fmt.Println(traceID, "[ERROR] AsrCloudMinds Client stream.Send error: ", err)
		}
	}()

	// 收数据流
	var serverInfo, message string
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(traceID, "[ERROR] AsrCloudMinds Client stream.Recv error: ", err)
			break
		}
		// 保存服务版本信息
		if res.Extend != "" {
			serverInfo = res.Extend
		}
		// 打印返回值
		if res.Message != "" {
			message = res.Message
			fmt.Println(traceID, "[INFO] AsrCloudMinds Client resMessage: ", message)
		}
	}
	err = stream.CloseSend()

	if err != nil {
		fmt.Println(traceID, "[ERROR] AsrCloudMinds Client close stream error: ", err)
	}
	fmt.Println(traceID, "[INFO] AsrCloudMinds Client serverInfo: ", serverInfo)
	fmt.Println(traceID, "[INFO] AsrCloudMinds Final Result is: ", message)
	return message
}

func asrControlCallOnce(addr, agentID, path string) string {
	traceID := uuid.New().String()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(traceID, "[ERROR] AsrControl Client grpc.Dial error: ", err)
		return err.Error()
	}
	defer conn.Close()

	c := asrcontrol.NewSpeechClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	stream, err := c.StreamingRecognize(ctx)
	if err != nil {
		fmt.Println(traceID, "[ERROR] AsrControl Client c.StreamingRecognize error: ", err)
		return err.Error()
	}
	aqa := asrcontrol.RecognitionRequest{
		CommonReqInfo: &common.CommonReqInfo{
			Guid:        traceID,
			Timestamp:   time.Now().Unix(),
			Version:     "1.0",
			TenantId:    traceID,
			UserId:      traceID,
			RobotId:     traceID,
			ServiceCode: "ginger",
		},
		Body: &asrcontrol.Body{
			Option: map[string]string{
				"returnDetail":  "true",
				"recognizeOnly": "true",
				"tstAgentId":    agentID,
			},
			Data: &asrcontrol.Body_Data{
				Rate:     16000,
				Format:   "pcm",
				Language: "CH",
				Dialect:  "zh",
				Vendor:   "CloudMinds",
			},
			NeedWakeup: false,
			Type:       1,
		},
		Extra: nil,
	}
	audioFile, err := os.Open(path)
	if err != nil {
		fmt.Println(traceID, "[ERROR] AsrControl Client Open File error: ", err)
		return err.Error()
	}
	status := 0
	var buffer = make([]byte, 1280)
	skip := false
	if strings.Contains(path, ".wav") {
		skip = true
	}
	cnt := 0

	for {
		lens, err := audioFile.Read(buffer)
		if err != nil || lens == 0 {
			status = 2
		}
		if skip {
			aqa.Body.Data.Speech = buffer[44:lens]
			skip = false
		} else {
			aqa.Body.Data.Speech = buffer[:lens]
		}
		cnt++
		if cnt == 40 {
			aqa.Body.StreamFlag = 2
			time.Sleep(time.Duration(200) * time.Millisecond)
		}

		// 向服务端发送 指令
		if err := stream.Send(&aqa); err != nil {
			fmt.Println(traceID, "[ERROR] AsrControl Client stream.Send error: ", err)
			return err.Error()
		}
		if status == 2 {
			break
		}
		status = 1
	}

	aqa.Body.Data.Speech = nil
	aqa.Extra = &common.Extra{ExtraType: "audioExtra", ExtraBody: "val"}
	// 向服务端发送 指令
	if err := stream.Send(&aqa); err != nil {
		fmt.Println(traceID, "[ERROR] AsrControl Client stream.Send error: ", err)
		return err.Error()
	}

	audioFile.Close()
	stream.CloseSend()
	// 接收 从服务端返回的数据流
	var message, s3path string
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if res.DetailMessage != nil {
			message = res.DetailMessage.Fields["recognizedText"].GetStringValue()
			fmt.Println(traceID, "[INFO] AsrControl Client resMessage: ", message)
			s3path = res.DetailMessage.Fields["wavPath"].GetStringValue()
		}
	}

	fmt.Println(traceID, "[INFO] AsrControl Client S3 WAV path is: ", s3path)
	fmt.Println(traceID, "[INFO] AsrControl Client Final Result is: ", message)
	return message
}

func Test_asrDebug(t *testing.T) {
	env := map[int]map[string]string{
		85: {
			"ConnAddrASRCloudMinds": "172.16.23.85:30591",
			"ConnAddrASRControl":    "10.155.254.85:20217",
		},
		86: {
			"ConnAddrASRCloudMinds": "172.16.23.85:30890",
			"ConnAddrASRControl":    "172.16.31.96:31571",
		},
		87: {},
		251: {
			"ConnAddrASRCloudMinds": "172.16.31.96:31571",
			"ConnAddrASRControl":    "172.16.31.28:31229",
		},
		231: {
			"ConnAddrASRCloudMinds": "172.16.31.96:31571",
			"ConnAddrASRControl":    "172.16.31.28:31229",
		},
		232: {
			"ConnAddrASRCloudMinds": "172.16.31.96:31571",
			"ConnAddrASRControl":    "172.16.31.96:32489",
		},
	}

	address := env[85]
	agentID := "2259"
	path := "D:\\GitLab\\develop\\gocrontask\\upload\\wav\\20220918_864972049990745_20220918095018284.wav"

	message2 := asrControlCallOnce(address["ConnAddrASRControl"], agentID, path)
	message1 := asrCloudMindsCallOnce(address["ConnAddrASRCloudMinds"], agentID, path)
	fmt.Println("AsrCloudMinds Final Result is", message1)
	fmt.Println("AsrControl Final Result is", message2)
}
