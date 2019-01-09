package nft

//API defines nft api
type API interface {
	ApplyFromFile(cfg string) error
	Apply(nft Nft) error
	DropRules(sub Nft) error
	Drop(family Family, table, chain string, handle int) error
	Get() (Nft, error)

	IPv4Set(family Family, table string, name string, ips ...string) error
	IPv4SetDel(family Family, table, name string, ips ...string) error
}
