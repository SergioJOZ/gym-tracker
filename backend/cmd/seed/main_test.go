package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapExerciseJSON(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "0001",
		"name": "Bench Press",
		"category": "strength",
		"body_part": "chest",
		"equipment": "barbell",
		"instructions": {
			"en": "Compound chest exercise"
		},
		"muscle_group": "chest",
		"secondary_muscles": ["triceps", "shoulders"],
		"target": "pectorals",
		"image": "images/0001-test.jpg",
		"gif_url": "videos/0001-test.gif",
		"media_id": "test123"
	}`)

	ex, err := mapExerciseJSON(raw, nil)
	require.NoError(t, err)
	assert.Equal(t, "Bench Press", ex.NameByLang["en"])
	assert.Equal(t, "Compound chest exercise", ex.DescriptionsByLang["en"])
	assert.Equal(t, "chest", ex.MuscleGroup)
	assert.Equal(t, "barbell", ex.Equipment)
	assert.Equal(t, "beginner", ex.Difficulty)
	assert.Equal(t, "strength", ex.Category)
	assert.Equal(t, "videos/0001-test.gif", ex.GIFPath)
	assert.Equal(t, "images/0001-test.jpg", ex.ThumbnailPath)
}

func TestMapExerciseJSON_WithTranslations(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "0001",
		"name": "Bench Press",
		"muscle_group": "chest"
	}`)

	translations := map[string]string{"0001": "Press de banca"}
	ex, err := mapExerciseJSON(raw, translations)
	require.NoError(t, err)
	assert.Equal(t, "Bench Press", ex.NameByLang["en"])
	assert.Equal(t, "Press de banca", ex.NameByLang["es"])
}

func TestMapExerciseJSON_Defaults(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "0002",
		"name": "Push Up",
		"muscle_group": "chest"
	}`)

	ex, err := mapExerciseJSON(raw, nil)
	require.NoError(t, err)
	assert.Equal(t, "Push Up", ex.NameByLang["en"])
	assert.Equal(t, "chest", ex.MuscleGroup)
	assert.Equal(t, "", ex.DescriptionsByLang["en"])
	assert.Equal(t, "", ex.Equipment)
	assert.Equal(t, "beginner", ex.Difficulty)
	assert.Equal(t, "", ex.Category)
}

func TestMapExerciseJSON_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{invalid json}`)

	_, err := mapExerciseJSON(raw, nil)
	assert.Error(t, err)
}

func TestMapExerciseJSON_MissingRequiredFields(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "550e8400-e29b-41d4-a716-446655440000"
	}`)

	_, err := mapExerciseJSON(raw, nil)
	assert.Error(t, err)
}
