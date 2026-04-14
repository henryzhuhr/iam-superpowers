package email

import (
	"fmt"

	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	gomail "gopkg.in/mail.v2"
)

type Service struct {
	cfg config.SMTPConfig
}

func New(cfg config.SMTPConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) SendVerificationCode(to string, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your verification code")
	m.SetBody("text/html", fmt.Sprintf(`
		<h1>Your verification code</h1>
		<p>Code: <strong>%s</strong></p>
		<p>This code expires in 5 minutes.</p>
	`, code))

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.User, s.cfg.Password)
	return d.DialAndSend(m)
}
