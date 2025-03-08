package dto

import "github.com/samcharles93/cinea/internal/entity"

type MovieDTO struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

func MovieToDTO(movie *entity.Movie) *MovieDTO {
	return &MovieDTO{
		ID:    movie.ID,
		Title: movie.Title,
	}
}

func MoviesToDTO(movies []*entity.Movie) []*MovieDTO {
	movieDTOs := make([]*MovieDTO, len(movies))
	for i, movie := range movies {
		movieDTOs[i] = MovieToDTO(movie)
	}
	return movieDTOs
}
