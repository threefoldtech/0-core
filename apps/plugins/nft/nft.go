package nft

//API defines nft api
type API interface {
	Apply(nft Nft) error
	Drop(family Family, table, chain string, handle int) error
	Find(filter ...Filter) ([]FilterRule, error)

	IPv4Set(family Family, table string, name string, ips ...string) error
	IPv4SetDel(family Family, table, name string, ips ...string) error
}
