package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	errImageNotFound    = errors.New("image not found")
	errItemNotFound     = errors.New("item not found")
	errCategoryNotFound = errors.New("category not found")
)

type Item struct {
	ID         int    `db:"id" json:"-"`
	Name       string `db:"name" json:"name"`
	Category   string `db:"category" json:"category"`
	CategoryID int    `db:"category_id" json:"-"`
	ImageName  string `db:"image_name" json:"image_name"`
}

type Category struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	List(ctx context.Context) ([]*Item, error)
	Select(ctx context.Context, id int) (*Item, error)
	SearchByKeyword(ctx context.Context, keyword string) ([]*Item, error)
}

type CategoryRepository interface {
	GetOrCreate(ctx context.Context, name string) (int, error)
	GetByID(ctx context.Context, id int) (*Category, error)
}

type itemRepository struct {
	db *sql.DB
}

type categoryRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) ItemRepository {
	return &itemRepository{db: db}
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	const query = `INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)`
	result, err := i.db.ExecContext(ctx, query, item.Name, item.CategoryID, item.ImageName)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	id, err := result.LastInsertId()
	if err == nil {
		item.ID = int(id)
	}
	return nil
}

func (i *itemRepository) List(ctx context.Context) ([]*Item, error) {
	const query = `
        SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
        FROM items i
        JOIN categories c ON i.category_id = c.id
    `
	rows, err := i.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var it Item
		var categoryName string
		if err := rows.Scan(&it.ID, &it.Name, &it.CategoryID, &categoryName, &it.ImageName); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		it.Category = categoryName
		items = append(items, &it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}
	return items, nil
}

// Select retrieves an item by id.
func (i *itemRepository) Select(ctx context.Context, id int) (*Item, error) {
	const query = `
        SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
        FROM items i
        JOIN categories c ON i.category_id = c.id
        WHERE i.id = ?
    `
	row := i.db.QueryRowContext(ctx, query, id)

	var it Item
	var categoryName string
	if err := row.Scan(&it.ID, &it.Name, &it.CategoryID, &categoryName, &it.ImageName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errItemNotFound
		}
		return nil, fmt.Errorf("failed to scan selected item: %w", err)
	}
	it.Category = categoryName
	return &it, nil
}

// SearchByKeyword searches items containing the keyword.
func (i *itemRepository) SearchByKeyword(ctx context.Context, keyword string) ([]*Item, error) {
	const query = `
        SELECT i.id, i.name, i.category_id, c.name as category_name, i.image_name 
        FROM items i
        JOIN categories c ON i.category_id = c.id
        WHERE i.name LIKE ?
    `
	searchPattern := "%" + keyword + "%"

	rows, err := i.db.QueryContext(ctx, query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var it Item
		var categoryName string
		if err := rows.Scan(&it.ID, &it.Name, &it.CategoryID, &categoryName, &it.ImageName); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		it.Category = categoryName
		items = append(items, &it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}
	return items, nil
}

func (c *categoryRepository) GetOrCreate(ctx context.Context, name string) (int, error) {
	var id int
	query := `SELECT id FROM categories WHERE name = ?`
	err := c.db.QueryRowContext(ctx, query, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("failed to query category: %w", err)
	}

	result, err := c.db.ExecContext(ctx, `INSERT INTO categories (name) VALUES (?)`, name)
	if err != nil {
		return 0, fmt.Errorf("failed to insert category: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(lastID), nil
}

// GetByID retrieves a category by id.
func (c *categoryRepository) GetByID(ctx context.Context, id int) (*Category, error) {
	var category Category
	query := `SELECT id, name FROM categories WHERE id = ?`
	err := c.db.QueryRowContext(ctx, query, id).Scan(&category.ID, &category.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errCategoryNotFound
		}
		return nil, fmt.Errorf("failed to scan category: %w", err)
	}
	return &category, nil
}

// StoreImage stores an image file.
func StoreImage(fileName string, image []byte) error {
	if err := os.WriteFile(fileName, image, 0644); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}
	return nil
}
