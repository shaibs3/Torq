package lookup

type LookupProvider interface {
	Lookup(ip string) (city string, country string, err error)
}
