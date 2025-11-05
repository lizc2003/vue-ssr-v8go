package alarm

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN

var gFeishuRobot *RobotFeiShu

func NewDefaultRobot(env, url, secret string) {
	gFeishuRobot = NewRobotFeiShu(env, defs.App, url, secret, "", "")
}

func SendAlert(msg string) {
	if gFeishuRobot != nil {
		gFeishuRobot.SendMsg(msg)
	}
}

type RobotFeiShu struct {
	Env        string
	url        string
	secret     string
	httpClient *http.Client
	bDisable   bool
}

func NewRobotFeiShu(env string, app string, url string, secret string, gitVersion, gitBranchName string) *RobotFeiShu {
	host, _ := os.Hostname()
	r := &RobotFeiShu{
		Env: "env: " + env + "\nhost: " + host + "\napp: " + app +
			"\n\n",
		url:        url,
		secret:     secret,
		httpClient: util.NewHttpClient(false),
	}

	return r
}

func (this *RobotFeiShu) SendMsg(msg string) error {
	if this.bDisable {
		return nil
	}

	var timestamp string
	var sign string
	now := time.Now()
	if this.secret != "" {
		timestamp = strconv.FormatInt(now.Unix(), 10)
		strSign := timestamp + "\n" + this.secret

		h := hmac.New(sha256.New, util.UnsafeStr2Bytes(strSign))
		h.Write([]byte{})
		sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
	}

	type msgData struct {
		Timestamp string `json:"timestamp,omitempty"`
		Sign      string `json:"sign,omitempty"`
		MsgType   string `json:"msg_type"`
		Content   struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	b := strings.Builder{}
	b.WriteString(this.Env)
	b.WriteString("time: ")
	b.WriteString(util.FormatTime(now))
	b.WriteByte('\n')
	b.WriteString(msg)

	data := msgData{Timestamp: timestamp, Sign: sign, MsgType: "text"}
	data.Content.Text = b.String()

	var resp string
	err := util.HttpPost(this.httpClient, this.url, nil, data, &resp)
	if err != nil {
		tlog.Errorf("feishu alert msg err: %s", err)
		return err
	}
	return nil
}

func (this *RobotFeiShu) SendCardMsg(card Card) error {
	if this.bDisable {
		return nil
	}

	msg := fsMessageV2{
		MsgType: "interactive",
		Card:    card,
	}
	msg.GetSign(this.secret)

	var respBody string
	if err := util.HttpPost(this.httpClient, this.url, nil, msg, &respBody); err != nil {
		tlog.Errorf("feishu alert card msg err: %s", err)
		return err
	}

	// 格式参考: {"code":19001,"data":{},"msg":"param invalid: incoming webhook access token invalid"}
	var resp struct {
		StatusCode    int    `json:"code"`
		StatusMessage string `json:"msg"`
	}
	if err := json.Unmarshal(util.UnsafeStr2Bytes(respBody), &resp); err != nil {
		tlog.Errorf("feishu alert card msg err: %s", err)
		return err
	}

	if resp.StatusCode != 0 {
		tlog.Errorf("feishu alert card msg err: %s", resp.StatusMessage)
		return errors.New(resp.StatusMessage)
	}
	return nil
}

// -----------------------------------------------------------------------------

type Conf struct {
	WideScreenMode bool `json:"wide_screen_mode"`
	EnableForward  bool `json:"enable_forward"`
}

type Te struct {
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type Element struct {
	Tag      string    `json:"tag"`
	Text     Te        `json:"text"`
	Content  string    `json:"content"`
	Elements []Element `json:"elements"`
}

type Titles struct {
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type Headers struct {
	Title    Titles `json:"title"`
	Template string `json:"template"`
}

type Card struct {
	Config   Conf      `json:"config"`
	Elements []Element `json:"elements"`
	Header   Headers   `json:"header"`
}

type fsMessageV2 struct {
	Timestamp string `json:"timestamp,omitempty"`
	Sign      string `json:"sign,omitempty"`
	MsgType   string `json:"msg_type"`
	Email     string `json:"email"` //@所使用字段
	Card      Card   `json:"card"`
}

func (this *fsMessageV2) GetSign(secret string) {
	now := time.Now()
	if secret != "" {
		timestamp := strconv.FormatInt(now.Unix(), 10)
		strSign := timestamp + "\n" + secret
		h := hmac.New(sha256.New, util.UnsafeStr2Bytes(strSign))
		h.Write([]byte{})
		this.Sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
		this.Timestamp = timestamp
	}
}
