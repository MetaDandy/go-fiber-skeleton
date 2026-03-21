package mail

// EmailRequest normaliza requests entre proveedores
type EmailRequest struct {
	To       string
	From     string
	FromName string
	Subject  string
	Html     string
	Text     string
	Token    string
	Link     string
}

// EmailResponse normaliza respuestas entre proveedores
type EmailResponse struct {
	ID     string
	Status string
	Error  error
}
