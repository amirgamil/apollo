package backend

import (
	"time"

	readability "github.com/go-shiori/go-readability"
)

func Scrape(link string) (Data, error) {
	// initRegeEx()
	// resp, err := http.Get(link)
	// if err != nil {
	// 	fmt.Println("Error trying to scrape writing of the post")
	// }
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading the body: ", link, " while scraping")
	// }
	// //load the HTML document
	// doc, err := goquery.NewDocumentFromReader(body)
	// if err != nil {
	// 	fmt.Println("Error creating a document from the queried website!")
	// }
	// convertToParagraphs(doc)
	// return Data{}
	article, err := readability.FromURL(link, 30*time.Second)
	if err != nil {
		return Data{}, err
	}
	return Data{Title: article.Title, Link: link, Content: article.TextContent, Tags: make([]string, 0)}, nil
}

// //Much of this code is ported/inspired from the excellent mercury-parser library
// func getData(body string) {
// 	title := getTitle(body)

// 	content := getContent(body)
// }

// func getTitle(body string) {
// 	//do stuff
// }

// func getContent(body string) {
// 	//do stuff
// }

// //helper methods to convert p-like methods to p tags
// func convertToParagraphs(document *goquery.Document) {
// 	brsToPs(document)
// 	divsToPs(document)
// 	spansToPs(document)
// }

// const BLOCK_LEVEL_TAGS = []string{"article", "aside", "blockquote", "body", "br", "button", "canvas", "caption", "col", "colgroup", "dd", "div", "dl", "dt", "embed", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hgroup", "hr", "li", "map", "object", "ol", "output", "p", "pre", "progress", "section", "table", "tbody", "textarea", "tfoot", "th", "thead", "tr", "ul", "video"}

// var re *regexp.Regexp

// //initialize necessary regex variables
// func initRegeEx() {
// 	var err error
// 	re, err = regexp.Compile("^(" + strings.Join(BLOCK_LEVEL_TAGS, "|") + ")$")
// 	if err != nil {
// 		fmt.Println("Error parsing the block level tags regex")
// 	}
// }

// //converts consecutive <br /> tags into <p /> tags e.g. <br /><br />
// func brsToPs(document *goquery.Document) {
// 	collapsing := false
// 	document.Find("br").Each(func(i int, s *goquery.Selection) {
// 		element := s.Get(0)
// 		//get the immediate sibling
// 		nextElement := s.Next().Get(0)

// 		//if the next element is also a br, we want to remove current one and flag on next iteration to
// 		//potentially convert the "current" nextElement (on next iteration will be current) that we may have to convert to text
// 		if nextElement != nil && strings.ToLower(nextElement.Data) {
// 			collapsing = true
// 			s.Remove()
// 		} else if collapsing {
// 			//we've reached the last <br /> tag in a serious of them, reset the flag and convert the child text into a p text
// 			collapsing = false
// 			convertElementToP(element, document, true)
// 		}
// 	})
// }

// func divsToPs(document *goquery.Document) {

// }

// func spansToPs(document *goquery.Document) {

// }

// //helper method which takes a DOM node and converts it into a p element
// //node: element to remove
// //dom: document level container for DOM manipulation
// func convertElementToP(node *html.Node, dom *goquery.Document, isBr bool) {
// 	if isBr {
// 		sibling := node.NextSibling
// 		//new p element we are creating that will replace the text
// 		newPElement := html.Node{Type: html.ElementNode, Data: "p"}
// 		//we will keep appending the text to our newPElement while we don't encounter a block element (which adds a new line)
// 		for sibling != nil && !(sibling.Data && re.MatchString(sibling.Data)) {
// 			newPElement.AppendChild(sibling)
// 			sibling = sibling.NextSibling
// 		}

// 		//replace node with our new p element
// 		replaceWith(node, &newPElement)
// 	}
// }

// //helper method to replace a node with another node
// func replaceWith(old *html.Node, new *html.Node) {
// 	parent := old.Parent
// 	if parent != nil {
// 		parent.RemoveChild(old)
// 		parent.AppendChild(new)
// 	}
// }
