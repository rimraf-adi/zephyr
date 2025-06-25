package netutil

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// HTMLParser represents an HTML parser for PyPI simple index
type HTMLParser struct {
	doc *html.Node
}

// NewHTMLParser creates a new HTML parser
func NewHTMLParser(htmlContent string) (*HTMLParser, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	return &HTMLParser{doc: doc}, nil
}

// ParsePyPISimpleIndex parses a PyPI simple index HTML page
func ParsePyPISimpleIndex(htmlContent string) ([]string, error) {
	parser, err := NewHTMLParser(htmlContent)
	if err != nil {
		return nil, err
	}
	
	return parser.ExtractPackageLinks()
}

// ExtractPackageLinks extracts package links from PyPI simple index
func (p *HTMLParser) ExtractPackageLinks() ([]string, error) {
	var links []string
	
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					// Extract package name from href
					href := attr.Val
					if strings.HasSuffix(href, "/") {
						href = strings.TrimSuffix(href, "/")
					}
					if href != "" {
						links = append(links, href)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(p.doc)
	return links, nil
}

// ExtractDownloadLinks extracts download links from a package page
func (p *HTMLParser) ExtractDownloadLinks() ([]DownloadLink, error) {
	var links []DownloadLink
	
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href, text string
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
				}
			}
			
			// Extract text content
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				text = strings.TrimSpace(n.FirstChild.Data)
			}
			
			if href != "" && text != "" {
				// Check if it's a download link (ends with .whl, .tar.gz, etc.)
				if strings.HasSuffix(href, ".whl") || 
				   strings.HasSuffix(href, ".tar.gz") || 
				   strings.HasSuffix(href, ".zip") {
					links = append(links, DownloadLink{
						URL:  href,
						Text: text,
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(p.doc)
	return links, nil
}

// DownloadLink represents a download link from PyPI
type DownloadLink struct {
	URL  string
	Text string
}

// FetchAndParseHTML fetches HTML content and parses it
func FetchAndParseHTML(client *http.Client, url string) (*HTMLParser, error) {
	req, err := CreatePyPIRequest("GET", url)
	if err != nil {
		return nil, err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	return NewHTMLParser(string(body))
}

// ExtractTextContent extracts text content from HTML nodes
func ExtractTextContent(n *html.Node) string {
	var text strings.Builder
	
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(n)
	return strings.TrimSpace(text.String())
}

// FindElementByTag finds the first element with the given tag name
func FindElementByTag(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}
	
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := FindElementByTag(c, tagName); found != nil {
			return found
		}
	}
	
	return nil
}

// FindElementByClass finds the first element with the given class name
func FindElementByClass(n *html.Node, className string) *html.Node {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, className) {
				return n
			}
		}
	}
	
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := FindElementByClass(c, className); found != nil {
			return found
		}
	}
	
	return nil
}

// GetAttribute gets the value of an attribute from an HTML node
func GetAttribute(n *html.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// ParsePyPIPackagePage parses a PyPI package page to extract metadata
func ParsePyPIPackagePage(htmlContent string) (*PyPIPackageInfo, error) {
	parser, err := NewHTMLParser(htmlContent)
	if err != nil {
		return nil, err
	}
	
	info := &PyPIPackageInfo{}
	
	// Extract package name from title
	titleNode := FindElementByTag(parser.doc, "title")
	if titleNode != nil {
		title := ExtractTextContent(titleNode)
		// Parse title like "Package Name · PyPI"
		if strings.Contains(title, "·") {
			parts := strings.Split(title, "·")
			if len(parts) > 0 {
				info.Name = strings.TrimSpace(parts[0])
			}
		}
	}
	
	// Extract description
	descNode := FindElementByClass(parser.doc, "package-description")
	if descNode != nil {
		info.Description = ExtractTextContent(descNode)
	}
	
	// Extract download links
	downloadLinks, err := parser.ExtractDownloadLinks()
	if err != nil {
		return nil, err
	}
	info.DownloadLinks = downloadLinks
	
	return info, nil
}

// PyPIPackageInfo represents package information extracted from HTML
type PyPIPackageInfo struct {
	Name          string
	Description   string
	DownloadLinks []DownloadLink
} 