package fabclient

func convertArrayOfStringsToArrayOfByteArrays(args []string) [][]byte {
	res := make([][]byte, 0, len(args))
	for _, arg := range args {
		res = append(res, []byte(arg))
	}
	return res
}
