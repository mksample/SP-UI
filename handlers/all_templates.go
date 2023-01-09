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

// TextSecrets is a struct for passing UiNodeText context secrets to a template
type textSecret struct {
	Id   int64
	Text string
}

// Marshals data from a text secret map
func (ts *textSecret) marshal(m map[string]interface{}) {
	ts.Id = int64(m["id"].(float64))
	ts.Text = m["text"].(string)
}

// Default template functions, added to all templates
func globalFuncMap() template.FuncMap {

	return template.FuncMap{
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
		"assetPath": func(fs hashfs.FS, name string) string {
			if strings.HasPrefix(name, "/") {
				log.Printf("Error assetPath called with name: '%s' should not start with '/'", name)
			}
			path := fs.HashName(name)
			if strings.HasPrefix(path, "/") {
				return path
			}
			return fmt.Sprintf("/%s", path)
		},
		"getTextSecrets": func(node kratos.UiNode) []textSecret {
			ts := []textSecret{}
			secrets := node.Attributes.UiNodeTextAttributes.Text.Context["secrets"]
			v := reflect.ValueOf(secrets)
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
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				log.Printf("invalid dict call")
				return nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					log.Printf("invalid dict call")
					return nil
				}
				dict[key] = values[i+1]
			}
			return dict
		},
		"toUiNodePartial": func(node kratos.UiNode) string {
			if _, ok := node.GetTypeOk(); !ok {
				log.Printf("Error toUiNodePartial called with unset node type")
			} else {
				if node.Type == "a" {
					return "ui_node_anchor"
				} else if node.Type == "img" {
					return "ui_node_image"
				} else if node.Type == "input" {

					switch node.Attributes.UiNodeInputAttributes.Type {
					case "hidden":
						return "ui_node_input_hidden"
					case "submit":
						return "ui_node_input_button"
					case "button":
						return "ui_node_input_button"
					case "checkbox":
						return "ui_node_input_checkbox"
					default:
						return "ui_node_input_default"
					}
				} else if node.Type == "script" {
					return "ui_node_script"
				} else if node.Type == "text" {
					return "ui_node_text"
				}
				return "ui_node_input_default"
			}
			return "ui_node_input_default"
		},
		"getNodeLabel": func(node kratos.UiNode) string {
			if _, ok := node.GetTypeOk(); !ok {
				log.Printf("Error getNodeLabel called with unset node type")
			} else {
				if node.Type == "a" {
					return node.Attributes.UiNodeAnchorAttributes.Title.Text
				} else if node.Type == "img" {
					return node.Meta.Label.Text
				} else if node.Type == "input" {
					if node.Attributes.UiNodeInputAttributes.HasLabel() {
						return node.Attributes.UiNodeInputAttributes.Label.Text
					}
				}
			}
			if node.Meta.HasLabel() {
				return node.Meta.Label.Text
			} else {
				return ""
			}
		},
		"onlyNodesGroups": func(nodes []kratos.UiNode, groups string) []kratos.UiNode {
			var filtered []kratos.UiNode
			sGroups := strings.Split(groups, ",")
			if len(sGroups) == 0 {
				return nodes
			}
			for _, n := range nodes {
				for _, fg := range sGroups {
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
