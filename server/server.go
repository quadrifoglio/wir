package server

var (
	NodeID byte
)

// Init initializes the parameters
// of the server
func Init(nodeId byte, database string) error {
	NodeID = nodeId
	return InitDatabase(database)
}
