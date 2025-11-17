package emoji

type Emoji struct {
	Char        string
	Name        string
	Category    string
	Subcategory string
	Keywords    []string
}

type EmojiCategory string

const (
	CategoryAll           EmojiCategory = "All"
	CategorySmileysPeople EmojiCategory = "Smileys & People"
	CategoryAnimalsNature EmojiCategory = "Animals & Nature"
	CategoryFoodDrink     EmojiCategory = "Food & Drink"
	CategoryTravelPlaces  EmojiCategory = "Travel & Places"
	CategoryActivities    EmojiCategory = "Activities"
	CategoryObjects       EmojiCategory = "Objects"
	CategorySymbols       EmojiCategory = "Symbols"
	CategoryFlags         EmojiCategory = "Flags"
)

func (c EmojiCategory) String() string {
	return string(c)
}
