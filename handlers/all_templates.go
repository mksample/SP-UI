package handlers

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"reflect"
	"strings"

	"github.com/benbjohnson/hashfs"
	kratos "github.com/ory/kratos-client-go"
)

// Each HTML file has an associated variable that is populated
// with its content at build time - see https://golang.org/pkg/embed/
//
// This means the binary contains everything it needs to serve the site
//
var (
	// Shared templates
	//
	//go:embed layout.html
	layoutTemplate string
	//go:embed partials/messages.html
	messagesTemplate string
	//go:embed partials/ui.html
	uiTemplate string
	//go:embed partials/ui_nodes.html
	uiNodesTemplate string
	//go:embed partials/ui_node_text.html
	uiNodeTextTemplate string
	//go:embed partials/ui_node_script.html
	uiNodeScriptTemplate string
	//go:embed partials/ui_node_input_hidden.html
	uiNodeInputHiddenTemplate string
	//go:embed partials/ui_node_input_default.html
	uiNodeInputDefaultTemplate string
	//go:embed partials/ui_node_input_checkbox.html
	uiNodeInputCheckboxTemplate string
	//go:embed partials/ui_node_input_button.html
	uiNodeInputButtonTemplate string
	//go:embed partials/ui_node_image.html
	uiNodeImageTemplate string
	//go:embed partials/ui_node_anchor.html
	uiNodeAnchorTemplate string
	//go:embed partials/ui_docs_button.html
	uiDocsButtonTemplate string
	//go:embed partials/ui_screen_button.html
	uiScreenButtonTemplate string
	//go:embed partials/fork_me.html
	forkMeTemplate string

	// Template per page
	//
	//go:embed login.html
	loginTemplate string
	//go:embed recovery.html
	recoveryTemplate string
	//go:embed registration.html
	registrationTemplate string
	//go:embed settings.html
	settingsTemplate string
	//go:embed verification.html
	verificationTemplate string
	//go:embed welcome.html
	welcomeTemplate string
	//go:embed error.html
	errorTemplate string

	emptyFuncMap         = template.FuncMap{}
	emptyStmulusTemplate = `
	{{define "stimulus"}}
	<!-- Empty stimulus template -->
	{{end}}`
)

// TemplateName provides a typesafe way of referring to templates
type TemplateName string

const (
	loginPage        = TemplateName("login")
	recoveryPage     = TemplateName("recovery")
	registrationPage = TemplateName("registration")
	settingsPage     = TemplateName("settings")
	verificationPage = TemplateName("verification")
	welcomePage      = TemplateName("welcome")
	errorPage        = TemplateName("error")
)

// Register all the Templates during initialisation
func init() {
	type tmpl struct {
		name      TemplateName     // Template name that handler code will refer to - one for each 'page'
		fmap      template.FuncMap // List of functions used inside the template
		templates []string         // List of HTML templates, snippets etc that make up the page
		stimulus  string           // Optional stimulus controller code
	}
	// All pages get the commonTemplates
	commonTemplates := []string{
		layoutTemplate,
		messagesTemplate,
		uiTemplate,
		uiNodesTemplate,
		uiNodeTextTemplate,
		uiNodeScriptTemplate,
		uiNodeInputHiddenTemplate,
		uiNodeInputDefaultTemplate,
		uiNodeInputCheckboxTemplate,
		uiNodeInputButtonTemplate,
		uiNodeImageTemplate,
		uiNodeAnchorTemplate,
		uiDocsButtonTemplate,
		uiScreenButtonTemplate,
		forkMeTemplate,
	}

	// The templates and their associated functions to include etc
	templates := []tmpl{
		{name: loginPage, fmap: emptyFuncMap, templates: []string{loginTemplate}},
		{name: recoveryPage, fmap: emptyFuncMap, templates: []string{recoveryTemplate}},
		{name: registrationPage, fmap: emptyFuncMap, templates: []string{registrationTemplate}},
		{name: settingsPage, fmap: emptyFuncMap, templates: []string{settingsTemplate}},
		{name: verificationPage, fmap: emptyFuncMap, templates: []string{verificationTemplate}},
		{name: welcomePage, fmap: emptyFuncMap, templates: []string{welcomeTemplate}},
		{name: errorPage, fmap: emptyFuncMap, templates: []string{errorTemplate}},
	}
	for _, t := range templates {
		stimulusTemplate := emptyStmulusTemplate
		if t.stimulus != "" {
			stimulusTemplate = t.stimulus
		}
		tmpl := append(commonTemplates, t.templates...)
		tmpl = append(tmpl, stimulusTemplate)

		// Ammend the global functions to the funcMap
		for k, v := range globalFuncMap() {
			t.fmap[k] = v
		}

		if err := RegisterTemplate(t.name, t.fmap, tmpl...); err != nil {
			// If we have a problem with a template, abort the app
			log.Fatalf("%v template error: %v", t.name, err)
		}
	}
}

// TextSecrets is a struct used for passing Lookup Secrets to a template.
// Ref: https://www.ory.sh/docs/kratos/mfa/lookup-secrets
type textSecret struct {
	Id   int64
	Text string
}

// Marshals data from a text secret map.
func (ts *textSecret) marshal(m map[string]interface{}) {
	ts.Id = int64(m["id"].(float64))
	ts.Text = m["text"].(string)
}

// Default template functions, added to all templates.
func globalFuncMap() template.FuncMap {

	return template.FuncMap{
		// Functions for returning safe HTML elements (URL, HTML attributes, JavaScript)
		// See https://pkg.go.dev/html/template#hdr-Contexts for more information
		"safeURL": func(s *string) template.URL {
			if s == nil {
				return ""
			}
			return template.URL(*s)
		},
		"safeAttr": func(s *string) template.HTMLAttr {
			if s == nil {
				return ""
			}
			return template.HTMLAttr(*s)
		},
		"safeJS": func(s *string) template.JS {
			if s == nil {
				return ""
			}
			return template.JS(*s)
		},

		// Returns a hashed path of the asset being used
		"assetPath": func(fs hashfs.FS, name string) string {
			if strings.HasPrefix(name, "/") {
				log.Printf("assetPath: called with name '%s', should not start with '/'", name)
			}
			path := fs.HashName(name)
			if strings.HasPrefix(path, "/") {
				return path
			}
			return fmt.Sprintf("/%s", path)
		},

		// Attempts to parse UI node text context into text secrets suitable for templates
		// See the type textSecret above for field names
		"getTextSecrets": func(node kratos.UiNode) []textSecret {
			secrets := node.Attributes.UiNodeTextAttributes.Text.Context["secrets"]
			v := reflect.ValueOf(secrets)
			ts := []textSecret{}

			if v.Kind() == reflect.Slice {
				for i := 0; i < v.Len(); i++ {
					m := v.Index(i).Interface().(map[string]interface{})
					newTs := textSecret{}
					newTs.marshal(m)
					ts = append(ts, newTs)
				}
				return ts
			}
			return nil
		},

		// Combines x number of keys and values into a map for template access
		//
		// Example: {{template "xyz" dict "Foo" "a" "Bar" .SomeValue}}
		// passes .Foo = "a" and .Bar = .SomeValue to the template xyz
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				log.Printf("dict: uneven number of keys and values for %v", values)
				return nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					log.Printf("dict: could not convert %v to string ", values[i])
					return nil
				}
				dict[key] = values[i+1]
			}
			return dict
		},

		// Returns the type of partial to display for a node
		"toUiNodePartial": func(node kratos.UiNode) string {
			if _, ok := node.GetTypeOk(); ok {
				switch node.Type {
				case "a":
					return "ui_node_anchor"
				case "img":
					return "ui_node_image"
				case "input":
					switch node.Attributes.UiNodeInputAttributes.Type {
					case "hidden":
						return "ui_node_input_hidden"
					case "submit":
						return "ui_node_input_button"
					case "button":
						return "ui_node_input_button"
					case "checkbox":
						return "ui_node_input_checkbox"
					}
				case "script":
					return "ui_node_script"
				case "text":
					return "ui_node_text"
				}
			}
			return "ui_node_input_default"
		},

		// Returns a node label based on the type of node passed
		"getNodeLabel": func(node kratos.UiNode) string {
			if _, ok := node.GetTypeOk(); ok {
				switch node.Type {
				case "a":
					return node.Attributes.UiNodeAnchorAttributes.Title.Text
				case "img":
					return node.Meta.Label.Text
				case "input":
					if node.Attributes.UiNodeInputAttributes.HasLabel() {
						return node.Attributes.UiNodeInputAttributes.Label.Text
					}
				}
			}

			// If no type given or no input label attempt to get from meta
			if node.Meta.HasLabel() {
				return node.Meta.Label.Text
			} else {
				return ""
			}
		},

		// Returns nodes with only the matching group type(s). If groups is blank all nodes are returned
		// Groups are specified with the format "groupa,groupb"
		"onlyNodesGroups": func(nodes []kratos.UiNode, groups string) []kratos.UiNode {
			var filtered []kratos.UiNode
			filterGroups := strings.Split(groups, ",")
			if len(filterGroups) == 0 {
				return nodes
			}

			for _, n := range nodes {
				for _, fg := range filterGroups {
					if n.Group == fg {
						filtered = append(filtered, n)
						break
					} else if fg == "all" {
						return nodes
					}
				}
			}
			return filtered
		},
	}
}
