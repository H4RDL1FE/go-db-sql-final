package main

import (
	"database/sql"
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

	// Устанавливаем идентификатор в ожидаемый объект для сравнения
	expectedParcel := parcel
	expectedParcel.Number = id

	// Получение
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	// Сравниваем структуры целиком
	require.Equal(t, expectedParcel, storedParcel)

	// Удаление
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверка удаления
	_, err = store.Get(id)
	require.Error(t, err) // Ожидаем ошибку, поскольку посылка должна быть удалена
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

	// Инициализация parcelMap
	parcelMap := make(map[int]Parcel)

	// Определяем клиента для тестирования
	clientID := rand.Intn(10000) // Генерация уникального идентификатора клиента

	// Создаем и добавляем тестовые посылки
	for i := 0; i < 3; i++ {
		testParcel := getTestParcel()
		testParcel.Client = clientID // Устанавливаем идентификатор клиента для тестовой посылки

		// Добавляем посылку в базу данных и в parcelMap
		id, err := store.Add(testParcel)
		require.NoError(t, err)
		testParcel.Number = id     // Обновляем номер посылки после добавления в БД
		parcelMap[id] = testParcel // Сохраняем посылку в map
	}

	// Получаем посылки по идентификатору клиента
	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err)

	// Проверяем, что каждая полученная посылка находится в parcelMap
	for _, sp := range storedParcels {
		ep, found := parcelMap[sp.Number]
		require.True(t, found, "Посылка с номером %d не найдена среди ожидаемых", sp.Number)

		// Проверяем совпадение всех полей
		require.Equal(t, ep, sp, "Полученная посылка не совпадает с ожидаемой")
	}
}
