package smtp

type Extension = string

const (
	SMTPExtAuth     Extension = "AUTH"
	SMTPExtStartTLS Extension = "STARTTLS"
)
