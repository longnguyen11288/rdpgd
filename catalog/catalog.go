package catalog

func Catalog() (string, error) {
	catalog := `
{
"services": [{
  "id": "rdpg-0.0.1",
  "name": "rdpg",
  "description": "A Relilable Distributed PostgrSQL Service",
  "bindable": true,
  "plans": [{
    "id": "rdpg-0.0.0.1-small",
    "name": "small",
    "description": "A small shared reliable PostgreSQL database."
  }]
}
}
`
return catalog,nil
}
