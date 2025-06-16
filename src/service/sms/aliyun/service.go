package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type AlibabaSMSService struct {
	appId     *string
	signature *string
	client    *dysmsapi20170525.Client
	signName  *string
}

func NewAlibabaSMSClient(accessKeyId string, accessKeySecret string) (*dysmsapi20170525.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String("dysmsapi.ap-southeast-1.aliyuncs.com")
	return dysmsapi20170525.NewClient(config)
}

func NewAlibabaSMSService(client *dysmsapi20170525.Client, signName *string) *AlibabaSMSService {
	return &AlibabaSMSService{
		client:   client,
		signName: signName,
	}
}

func (a *AlibabaSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers: a.toStringNumbersPtr(numbers),
		SignName:     a.signName,
	}

	argsMap := make(map[string]string, len(args))
	for k, arg := range args {
		argsMap[strconv.Itoa(k)] = arg
	}
	// 这意味着，你的模板必须是 你的短信验证码是{0}
	// 你的短信验证码是{code}
	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}

	req.TemplateParam = a.toStringPtr(string(bCode))
	req.TemplateCode = a.toStringPtr(tplId)

	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_, _err := a.client.SendSms(req)
		if _err != nil {
			return _err
		}
		return nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		fmt.Println(tea.StringValue(error.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			fmt.Println(recommend)
		}
		_, _err := util.AssertAsString(error.Message)
		if _err != nil {
			return _err
		}
	}

	return nil
}

func (t *AlibabaSMSService) toStringPtr(str string) *string {
	return &str
}

func (t *AlibabaSMSService) toStringNumbersPtr(numbers []string) *string {
	// 阿里云多个手机号中间用英文逗号隔开
	data := strings.Join(numbers, ",")
	return &data
}
