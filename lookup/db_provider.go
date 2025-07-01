package lookup

type DbProvider interface {
	Lookup(ip string) (city string, country string, err error)
}
