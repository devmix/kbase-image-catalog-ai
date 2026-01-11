package api

import (
	"sort"
)

// sortCatalogs sorts catalogs based on specified criteria
func SortCatalogs(catalogs []map[string]interface{}, sortBy, sortOrder string) []map[string]interface{} {
	// Default to sorting by name ascending if no parameters are provided
	if sortBy == "" {
		sortBy = "name"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	// Create a copy of the slice to avoid modifying original
	sortedCatalogs := make([]map[string]interface{}, len(catalogs))
	copy(sortedCatalogs, catalogs)

	switch sortBy {
	case "name":
		if sortOrder == "desc" {
			sort.SliceStable(sortedCatalogs, func(i, j int) bool {
				name1, _ := sortedCatalogs[i]["name"].(string)
				name2, _ := sortedCatalogs[j]["name"].(string)
				return name1 > name2
			})
		} else {
			sort.SliceStable(sortedCatalogs, func(i, j int) bool {
				name1, _ := sortedCatalogs[i]["name"].(string)
				name2, _ := sortedCatalogs[j]["name"].(string)
				return name1 < name2
			})
		}
	case "imageCount":
		if sortOrder == "desc" {
			sort.SliceStable(sortedCatalogs, func(i, j int) bool {
				count1, _ := sortedCatalogs[i]["imageCount"].(int)
				count2, _ := sortedCatalogs[j]["imageCount"].(int)
				return count1 > count2
			})
		} else {
			sort.SliceStable(sortedCatalogs, func(i, j int) bool {
				count1, _ := sortedCatalogs[i]["imageCount"].(int)
				count2, _ := sortedCatalogs[j]["imageCount"].(int)
				return count1 < count2
			})
		}
	case "lastUpdate":
		if sortOrder == "desc" {
			sort.SliceStable(sortedCatalogs, func(i, j int) bool {
				update1, _ := sortedCatalogs[i]["lastUpdate"].(string)
				update2, _ := sortedCatalogs[j]["lastUpdate"].(string)
				return update1 > update2
			})
		} else {
			sort.SliceStable(sortedCatalogs, func(i, j int) bool {
				update1, _ := sortedCatalogs[i]["lastUpdate"].(string)
				update2, _ := sortedCatalogs[j]["lastUpdate"].(string)
				return update1 < update2
			})
		}
	default:
		// Default to name sorting if an invalid sort parameter is provided
		sort.SliceStable(sortedCatalogs, func(i, j int) bool {
			name1, _ := sortedCatalogs[i]["name"].(string)
			name2, _ := sortedCatalogs[j]["name"].(string)
			return name1 < name2
		})
	}

	return sortedCatalogs
}

// sortCatalogImages sorts images in a catalog based on specified criteria
func SortCatalogImages(indexData map[string]interface{}, sortBy, sortOrder string) []map[string]interface{} {
	// Default to sorting by filename ascending if no parameters are provided
	if sortBy == "" {
		sortBy = "filename"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	if len(indexData) == 0 {
		return make([]map[string]interface{}, 0)
	}

	// Convert the map to an array for consistent sorting
	var images []map[string]interface{}
	for k, v := range indexData {
		if img, ok := v.(map[string]interface{}); ok {
			img["filename"] = k
			images = append(images, img)
		}
	}

	// Sort the array based on the specified criteria
	switch sortBy {
	case "shortName":
		if sortOrder == "desc" {
			sort.SliceStable(images, func(i, j int) bool {
				filename1, _ := images[i]["short_name"].(string)
				filename2, _ := images[j]["short_name"].(string)
				return filename1 > filename2
			})
		} else {
			sort.SliceStable(images, func(i, j int) bool {
				filename1, _ := images[i]["filename"].(string)
				filename2, _ := images[j]["filename"].(string)
				return filename1 < filename2
			})
		}
	case "description":
		if sortOrder == "desc" {
			sort.SliceStable(images, func(i, j int) bool {
				filename1, _ := images[i]["description"].(string)
				filename2, _ := images[j]["description"].(string)
				return filename1 > filename2
			})
		} else {
			sort.SliceStable(images, func(i, j int) bool {
				filename1, _ := images[i]["description"].(string)
				filename2, _ := images[j]["description"].(string)
				return filename1 < filename2
			})
		}
	// Add other sorting cases as needed
	default:
		// Default to filename sorting if an invalid sort parameter is provided
		sort.SliceStable(images, func(i, j int) bool {
			filename1, _ := images[i]["filename"].(string)
			filename2, _ := images[j]["filename"].(string)
			return filename1 < filename2
		})
	}

	return images
}
