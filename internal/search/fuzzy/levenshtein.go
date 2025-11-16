package fuzzy

// LevenshteinDistance calculates the edit distance between two strings
// Used for ranking when fuzzy match scores are equal
func LevenshteinDistance(s1, s2 string) int {
	s1Len := len(s1)
	s2Len := len(s2)

	// Create a matrix to store distances
	matrix := make([][]int, s1Len+1)
	for i := range matrix {
		matrix[i] = make([]int, s2Len+1)
	}

	// Initialize first row and column
	for i := 0; i <= s1Len; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= s2Len; j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= s1Len; i++ {
		for j := 1; j <= s2Len; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[s1Len][s2Len]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
