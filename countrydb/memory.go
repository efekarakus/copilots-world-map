package countrydb

type memoryDB struct {
	countries map[string]int
}

// NewMemoryDB returns a DB that stores the country visits in memory.
// If the database goes down, then all results are lost.
func NewMemoryDB() DB {
	return &memoryDB{
		countries: make(map[string]int),
	}
}

func (db *memoryDB) Save(country string) (int, error) {
	db.countries[country] = db.countries[country] + 1
	return db.countries[country], nil
}


func (db *memoryDB) Results() ([]Country, error) {
	countries := []Country{}
	for name, count := range db.countries {
		countries = append(countries, Country{
			Country: name,
			Visit:   count,
		})
	}
	return countries, nil
}

func (db *memoryDB) UniqueTotal() (int, error) {
	unique := 0
	for _, count := range db.countries {
		if count > 0 {
			unique += 1
		}
	}
	return unique, nil
}