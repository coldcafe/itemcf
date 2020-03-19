package itemcf

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RelatedMovie struct {
	Movie string
	Value float64
}

type ItemCF struct {
	SimMovieNum    int
	RecMovieNum    int
	TrainSet       map[string]map[string]float64
	TestSet        map[string]map[string]float64
	MovieSimMatrix map[string]map[string]float64
	MoviePopular   map[string]int
	MovieCount     int
}

func NewItemCF() *ItemCF {
	return &ItemCF{
		SimMovieNum:    10,
		RecMovieNum:    10,
		TrainSet:       map[string]map[string]float64{},
		TestSet:        map[string]map[string]float64{},
		MovieSimMatrix: map[string]map[string]float64{},
		MoviePopular:   map[string]int{},
	}
}

func (i *ItemCF) GetDataset(filename string) {
	pivot := 0.75
	trainSetLen := 0
	testSetLen := 0
	file, _ := os.Open(filename)
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		l, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineInfo := strings.Split(string(l), ",")
		user, movie, ratingStr := lineInfo[0], lineInfo[1], lineInfo[2]
		rating, _ := strconv.ParseFloat(ratingStr, 64)
		rand.Seed(time.Now().UnixNano())
		if rand.Float64() < pivot {
			if i.TrainSet[user] == nil {
				i.TrainSet[user] = map[string]float64{}
			}
			i.TrainSet[user][movie] = rating
			trainSetLen++
		} else {
			if i.TestSet[user] == nil {
				i.TestSet[user] = map[string]float64{}
			}
			i.TestSet[user][movie] = rating
			testSetLen++
		}
	}
	println("Split trainingSet and testSet success!")
	println("TrainSet = ", trainSetLen)
	println("TestSet = ", testSetLen)
}

func (i *ItemCF) CalcMovieSim() {
	for _, movies := range i.TrainSet {
		for movie := range movies {
			i.MoviePopular[movie]++
		}
	}
	i.MovieCount = len(i.MoviePopular)
	println("Total movie number = ", i.MovieCount)

	for _, movies := range i.TrainSet {
		for m1 := range movies {
			for m2 := range movies {
				if m1 != m2 {
					if i.MovieSimMatrix[m1] == nil {
						i.MovieSimMatrix[m1] = map[string]float64{}
					}
					i.MovieSimMatrix[m1][m2] += 1.0
				}
			}
		}
	}
	println("Build co-rated users matrix success!")

	for m1, relatedMovies := range i.MovieSimMatrix {
		for m2, count := range relatedMovies {
			if i.MoviePopular[m1] == 0 || i.MoviePopular[m2] == 0 {
				i.MovieSimMatrix[m1][m2] = 0
			} else {
				i.MovieSimMatrix[m1][m2] = count / math.Sqrt(float64(i.MoviePopular[m1]*i.MoviePopular[m2]))
			}
		}
	}
	println("Calculate movie similarity matrix success!")
}

func (i *ItemCF) Recommend(user string) []*RelatedMovie {
	K := i.SimMovieNum
	N := i.RecMovieNum
	rank := map[string]float64{}
	watchedMovies := i.TrainSet[user]
	for movie, rating := range watchedMovies {
		relatedMovies := mapSort(i.MovieSimMatrix[movie])
		if len(relatedMovies) > K {
			relatedMovies = relatedMovies[:K]
		}
		for _, relatedMovie := range relatedMovies {
			if watchedMovies[relatedMovie.Movie] == 0 {
				rank[relatedMovie.Movie] += relatedMovie.Value * rating
			}
		}
	}
	result := mapSort(rank)
	if len(result) > K {
		return result[:N]
	}
	return result
}

func (i *ItemCF) Evaluate() {
	println("Evaluating start ...")
	hit := 0
	recCount := 0
	testCount := 0

	allRecMovies := map[string]bool{}

	for user := range i.TrainSet {
		testMoives := i.TestSet[user]
		recMovies := i.Recommend(user)
		for _, rec := range recMovies {
			if testMoives[rec.Movie] != 0 {
				hit++
			}
			allRecMovies[rec.Movie] = true
		}
		recCount += len(recMovies)
		testCount += len(testMoives)
	}

	precision := float64(hit) / float64(1.0*recCount)
	recall := float64(hit) / float64(1.0*testCount)
	coverage := float64(len(allRecMovies)) / float64(1.0*i.MovieCount)

	fmt.Printf("precisioin=%.4f\trecall=%.4f\tcoverage=%.4f", precision, recall, coverage)
}

func mapSort(data map[string]float64) []*RelatedMovie {
	result := make([]*RelatedMovie, len(data))
	j := 0
	for m, v := range data {
		result[j] = &RelatedMovie{
			Movie: m,
			Value: v,
		}
		j++
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Value == result[j].Value {
			iName, _ := strconv.Atoi(result[i].Movie)
			jName, _ := strconv.Atoi(result[j].Movie)
			return iName < jName
		}
		return result[i].Value > result[j].Value
	})
	return result
}
