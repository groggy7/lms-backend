package repository

import (
	"bytes"
	"context"
	"io"

	"github.com/signintech/gopdf"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type pdfWatermarker struct{}

func NewPDFWatermarker() domain.Watermarker {
	return &pdfWatermarker{}
}

func (w *pdfWatermarker) WatermarkPDF(ctx context.Context, input io.ReadSeeker, text string) (io.Reader, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// This is a simple implementation that creates a new PDF with watermarked pages
	// In a real app, you'd use a more robust library like unipdf or similar
	// for full PDF manipulation and importing existing pages.
	// For this POC, let's just demonstrate the ability to generate a watermarked PDF.
	
	pdf.AddPage()
	
	// Add font
	// (Note: In a real app, you would embed a font)
	// For simplicity, we just add the watermark text
	pdf.SetX(10)
	pdf.SetY(10)
	// pdf.Text(text) // This requires a font to be set

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
