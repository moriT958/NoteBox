package models

var (
	nr *NoteRepository
)

func GetRepository() Repository {
	if nr == nil {
		panic("repository not initialized")
	}
	return nr
}
