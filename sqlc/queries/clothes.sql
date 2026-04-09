-- name: GetOneByCategoryStyleWeather :one
SELECT * FROM clothes
WHERE category = $1
  AND style = $2
  AND weather = $3
ORDER BY random()
LIMIT 1;
