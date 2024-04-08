package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupDatabase(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err)

	return db
}

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Получение
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, storedParcel.Client)
	require.Equal(t, parcel.Status, storedParcel.Status)
	require.Equal(t, parcel.Address, storedParcel.Address)

	// Удаление
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверка удаления
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление адреса
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Проверка
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление статуса
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// Проверка
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()

	store := NewParcelStore(db)

	client := randRange.Intn(10000) // Генерация случайного ID клиента
	var expectedParcels []Parcel

	// Добавление нескольких посылок для одного клиента
	for i := 0; i < 3; i++ {
		parcel := getTestParcel()
		parcel.Client = client
		id, err := store.Add(parcel)
		require.NoError(t, err)

		parcel.Number = id
		expectedParcels = append(expectedParcels, parcel)
	}

	// Получение посылок клиента
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(expectedParcels))

	// Проверка соответствия добавленных и полученных посылок
	for _, sp := range storedParcels {
		var found bool
		for _, ep := range expectedParcels {
			if sp.Number == ep.Number {
				require.Equal(t, ep.Client, sp.Client)
				require.Equal(t, ep.Status, sp.Status)
				require.Equal(t, ep.Address, sp.Address)
				found = true
				break
			}
		}
		require.True(t, found, fmt.Sprintf("Посылка с номером %d не найдена среди ожидаемых", sp.Number))
	}
}
