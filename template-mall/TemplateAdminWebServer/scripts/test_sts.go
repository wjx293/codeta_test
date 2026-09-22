//go:build ignore

package main

import (
	"fmt"
	"os"

	"template-mall/TemplateAdminWebServer/internal/config"
	"template-mall/TemplateAdminWebServer/internal/oss"
)

func main() {
	cfg := config.Load()
	issuer := oss.NewSTSIssuer(cfg)
	creds, err := issuer.Issue("templates/202608/test-sts-check.pptx")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK: token_len=%d expires=%s\n", len(creds.SecurityToken), creds.Expiration.Format("15:04:05"))
}
