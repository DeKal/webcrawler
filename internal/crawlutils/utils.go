package crawlutils

func GetUniqueSlide(slide []string) []string {
	set := make(map[string]bool)
	newSlide := make([]string, 0)

	for _, link := range slide {
		if _, ok := set[link]; !ok {
			set[link] = true
			newSlide = append(newSlide, link)
		}
	}
	return newSlide
}

func CombineMaps(map1 map[string][]string, map2 map[string][]string) map[string][]string {
	combinedMap := make(map[string][]string)
	for key, values := range map1 {
		combinedMap[key] = values
	}
	for key, values := range map2 {
		combinedMap[key] = values
	}
	return combinedMap
}
