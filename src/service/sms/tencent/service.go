package tencent

import (
	"context"
	"fmt"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type TencentSMSService struct {
	appId     *string
	signature *string
	client    *tencentsms.Client
}

func NewTencentSMSService(client *tencentsms.Client, appId string, signature string) *TencentSMSService {
	return &TencentSMSService{
		client:    client,
		appId:     ekit.ToPtr[string](appId),
		signature: ekit.ToPtr[string](signature),
	}
}

func (t *TencentSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := tencentsms.NewSendSmsRequest()
	req.SmsSdkAppId = t.appId
	req.SignName = t.signature
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.PhoneNumberSet = t.toStringPtrSlice(numbers)
	req.TemplateParamSet = t.toStringPtrSlice(args)
	resp, err := t.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("send sms failed %d, %s", status.Code, *status.Message)
		}
	}
	return nil
}

func (t *TencentSMSService) toStringPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data, func(idx int, src string) *string {
		return &src
	})
}
