package response

import "github.com/gofiber/fiber/v3"

// SendPDF writes a PDF document as an attachment download. fileName must be
// plain ASCII (it goes into the header unencoded).
func SendPDF(c fiber.Ctx, fileName string, content []byte) error {
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+fileName+`"`)
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	return c.Send(content)
}
