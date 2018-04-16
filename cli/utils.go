package cli

/*
 * deduplicate a slice of strings, keeping the order of the elements
 */
func deduplicate(input []string) []string {
	var output []string
	unique := map[string]interface{}{}
	for _, i := range input {
		unique[i] = new(interface{})
	}
	for _, i := range input {
		if _, ok := unique[i]; ok {
			output = append(output, i)
			delete(unique, i)
		}
	}
	return output
}
