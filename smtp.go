package main

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/bilireq"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"golang.org/x/sync/errgroup"
)

var bc *bilireq.Client

func init() {
	bp := os.Getenv("BEER_PROXY")
	if bp == "" {
		return
	}
	u := try.To1(url.Parse(bp))
	bc = bilireq.New(u.Hostname())
}

func initSMTPTry(app core.App) {
	if bc == nil {
		return
	}
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if args.SMTPAddr == "" {
			return e.Next()
		}
		ln, err := net.Listen("tcp", args.SMTPAddr)
		if err != nil {
			return err
		}
		slog.Warn("smtp server is running", "addr", ln.Addr())
		e.App.Store().Set("smtp-listener", ln)
		be := &Backend{app: e.App}
		srv := smtp.NewServer(be)
		srv.AllowInsecureAuth = true
		if e.App.IsDev() {
			srv.Debug = os.Stderr
		}
		go func() {
			err := srv.Serve(ln)
			if err != nil {
				slog.Error("stmp server listen failed", "error", err)
			}
		}()
		return e.Next()
	})
	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		ln, ok := e.App.Store().Get("smtp-listener").(net.Listener)
		if !ok || ln == nil {
			return e.Next()
		}
		ln.Close()
		return e.Next()
	})
}

type Backend struct {
	app  core.App
	auth sasl.Server
}

var _ smtp.Backend = (*Backend)(nil)

func (be *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	sess := &Session{app: be.app}
	sess.Reset()
	return sess, nil
}

type Session struct {
	app     core.App
	authSrv sasl.Server
	auth    *core.Record
	from    string
	outbox  []string // 需要转发出去
}

var _ smtp.AuthSession = (*Session)(nil)

func (sess *Session) Reset() {
	sess.outbox = []string{}
	sess.auth = nil
	sess.authSrv = sasl.NewPlainServer(func(identity, username, password string) error {
		username, _, _ = strings.Cut(username, "@")
		q := "application = {:user} && secret = {:pass}"
		p := dbx.Params{"user": username, "pass": password}
		ac, err := sess.app.FindFirstRecordByFilter(db.TableClients, q, p)
		if err != nil {
			return smtp.ErrAuthFailed
		}
		sess.auth = ac
		return nil
	})
}

func (sess *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}
func (sess *Session) Auth(mech string) (sasl.Server, error) {
	if mech != sasl.Plain {
		return nil, smtp.ErrAuthUnknownMechanism
	}
	return sess.authSrv, nil
}

var _ smtp.Session = (*Session)(nil)

func (sess *Session) Logout() error {
	return nil
}
func (sess *Session) Mail(from string, opts *smtp.MailOptions) error {
	if sess.auth == nil {
		return smtp.ErrAuthRequired
	}
	return nil
}
func (sess *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	if sess.auth == nil {
		return smtp.ErrAuthRequired
	}
	uid, domain, found := strings.Cut(to, "@")
	if !found {
		return ErrDomainNotFound
	}
	switch domain {
	case "bilibili.com":
		// don't need do anything
	case "live-open.bilibili.com":
		r, err := sess.app.FindFirstRecordByData(db.TableLinkeds, "openid", uid)
		if err != nil {
			return ErrUserNotFound
		}
		uid = r.GetString("uid")
	default:
		return ErrDomainNotFound
	}
	sess.outbox = append(sess.outbox, uid)
	return nil
}

func (sess *Session) Data(r io.Reader) (err error) {
	if sess.auth == nil {
		return smtp.ErrAuthRequired
	}

	app := sess.app
	logger := app.Logger()
	defer err0.Then(&err, nil, func() {
		logger.Error("发送B站消息出错", "error", err)
	})

	mr, _ := mail.CreateReader(r)
	subject, _ := mr.Header.Subject()
	var body string
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			ct, _ := try.To2(h.ContentType())
			if ct == "text/plain" {
				b := try.To1(io.ReadAll(p.Body))
				body = string(b)
			}
		}
	}
	application := sess.auth.GetString("bname")
	if application == "" {
		application = sess.auth.GetString("domain")
	}
	if application == "" {
		application = sess.auth.GetString("application")
	}
	msg := "应用: " + application
	msg += "\n主题: " + subject
	msg += "\n正文: \n" + body
	eg := new(errgroup.Group)
	for _, uidStr := range sess.outbox {
		eg.Go(func() error {
			id, err := strconv.Atoi(uidStr)
			if err != nil {
				return err
			}
			uid := int64(id)
			if _, err := bc.MsgSend2User(uid, msg); err != nil {
				return err
			}
			return nil
		})
	}
	err = eg.Wait()

	return err
}

var (
	ErrDomainNotFound = &smtp.SMTPError{
		Code:         550,
		EnhancedCode: smtp.EnhancedCode{5, 1, 2},
		Message:      "domain not found",
	}
	ErrUserNotFound = &smtp.SMTPError{
		Code:         550,
		EnhancedCode: smtp.EnhancedCode{5, 1, 1},
		Message:      "user unknown",
	}
	ErrUnauthorizedSender = &smtp.SMTPError{
		Code:         550,
		EnhancedCode: smtp.EnhancedCode{5, 7, 1},
		Message:      "Sender address rejected: not owned by authenticated user",
	}
)
