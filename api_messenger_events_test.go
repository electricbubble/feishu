package feishu

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestEvents(t *testing.T) {
	fsApp := testNewCustomApp()
	fsApp.opt.debug = true

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		_, err := io.Copy(&body, r.Body)
		requireNil(t, err)
		defer func() {
			_ = r.Body.Close()
		}()

		fmt.Println("####", body.String())

		eReq := new(eventRequest)
		err = json.Unmarshal(body.Bytes(), eReq)
		requireNil(t, err)
		logIndent(t, eReq)
		// json.NewDecoder(body).Decode(eReq)

		if eReq.Encrypt != "" {
			// 解密
			body, err = eReq.decrypt(fsApp.encryptKey)
			requireNil(t, err)
		}

		switch {
		case eReq.Type == EventTypeURLVerification:
			v := new(eventURLVerificationRequest)
			err := json.Unmarshal(body.Bytes(), v)
			requireNil(t, err)

			if v.Token != fsApp.verificationToken {
				requireNil(t, fmt.Errorf("unexpected event callback token (app: %s): %s", fsApp.verificationToken, v.Token))
				return
			}

			logIndent(t, v)

			resp := &eventURLVerificationResponse{Challenge: v.Challenge}
			bs, err := json.Marshal(resp)
			if err != nil {
				requireNil(t, err)
			}

			_, err = w.Write(bs)
			requireNil(t, err)
		case eReq.Schema == "2.0":
			if eReq.Header == nil {
				requireNil(t, errors.New("get event(2.0), but header is nil"))
				return
			}
			if eReq.Header.Token != fsApp.verificationToken {
				requireNil(t, fmt.Errorf("unexpected event callback token (app: %s): %s", fsApp.verificationToken, eReq.Header.Token))
				return
			}
			fmt.Println("事件类型:", eReq.Header.EventType)
			fmt.Println("事件:", string(eReq.Event))
		case eReq.EventHeaderV1.UUID != "":
			if eReq.EventHeaderV1.Token != fsApp.verificationToken {
				requireNil(t, fmt.Errorf("unexpected event callback token (app: %s): %s", fsApp.verificationToken, eReq.EventHeaderV1.Token))
				return
			}
			fmt.Println("事件类型:", eReq.EventHeaderV1.Type)
			fmt.Println("事件:", string(eReq.Event))
		}

		return

		fmt.Println("解密后", body.String())

		if eReq.Token != fsApp.verificationToken {
			log.Fatalln("unexpected event callback token", eReq)
		}

		fmt.Println("当前事件类型:", eReq.Type)

		if eReq.Type == EventTypeURLVerification {
			v := new(eventURLVerificationRequest)
			err := json.Unmarshal(body.Bytes(), v)
			requireNil(t, err)

			logIndent(t, v)

			resp := &eventURLVerificationResponse{Challenge: v.Challenge}
			bs, err := json.Marshal(resp)
			if err != nil {
				requireNil(t, err)
			}
			_, err = w.Write(bs)
			requireNil(t, err)
		}

		// logIndent(t, eReq)

		return

		expectedSignature := r.Header.Get("X-Lark-Signature")
		buf := new(bytes.Buffer)
		buf.WriteString(r.Header.Get("X-Lark-Request-Timestamp"))
		buf.WriteString(r.Header.Get("X-Lark-Request-Nonce"))
		buf.WriteString(fsApp.encryptKey)
		// body := new(bytes.Buffer)
		// _, err := io.Copy(body, r.Body)
		// requireNil(t, err)

		buf.Write(body.Bytes())

		fmt.Println("####", r.Header.Get("X-Lark-Request-Timestamp"))
		fmt.Println("####", r.Header.Get("X-Lark-Request-Nonce"))
		fmt.Println("####", r.Header.Get("X-Lark-Signature"))
		fmt.Println("####", body)

		h := sha256.New()
		h.Write(buf.Bytes())
		bs := h.Sum(nil)
		signature := fmt.Sprintf("%x", bs)

		fmt.Println("####", signature != expectedSignature)

		if signature != expectedSignature {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

	})

	err := http.ListenAndServe(":60081", nil)
	requireNil(t, err)
}

func Test_app_ListenEventCallback(t *testing.T) {
	fsApp := testNewCustomApp()
	fsApp.opt.debug = true

	fsApp.RegisterEventCallback(EventTypeMessageReceived, func(header EventHeaderV2, event json.RawMessage) {
		msg := new(EventMessageReceived)
		err := json.Unmarshal(event, msg)
		requireNil(t, err)

		logIndent(t, msg)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fsApp.ListenEventCallback(w, r)
	})

	err := http.ListenAndServe(":60081", nil)
	requireNil(t, err)
}
