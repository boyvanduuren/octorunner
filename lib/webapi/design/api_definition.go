package design

import (
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("octorunner", func() {
	Title("Octorunner status API")
	Description("A simple (read-only) API to query jobs ran by Octorunner")
	Version("1.0")
	Contact(func() {
		Name("B.C. van Duuren")
		Email("boy@vanduuren.xyz")
		URL("https://github.com/boyvanduuren/octorunner")
	})
	License(func() {
		Name("MIT")
		URL("https://github.com/boyvanduuren/octorunner/blob/master/LICENSE")
	})
	Docs(func() {
		Description("Setup guide")
		URL("https://github.com/boyvanduuren/octorunner/blob/master/README.md")
	})
	Scheme("http")
	Consumes("application/json")
	Produces("application/json")
	BasePath("/api")
})