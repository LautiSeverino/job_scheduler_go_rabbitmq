package repositories

import (
	"context"
	"fmt"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataStore struct {
	tx   pgx.Tx        // Transacción activa (nil si no hay transacción)
	pool *pgxpool.Pool // Conexión principal a la base de datos
}

func NewDataStore(pool *pgxpool.Pool) ports.IUnitOfWork {
	return &DataStore{
		pool: pool,
	}
}

// WithTx crea una nueva instancia de DataStore con una transacción
func (ds *DataStore) WithTx(tx pgx.Tx) *DataStore {
	return &DataStore{
		tx:   tx,
		pool: ds.pool,
	}
}

// GetTx devuelve la transacción activa (puede ser nil)
func (ds *DataStore) GetTx() pgx.Tx {
	return ds.tx
}

// Atomic ejecuta una función dentro de una transacción
func (ds *DataStore) Atomic(ctx context.Context, cb ports.FAtomicCallback) (err error) {
	// Si ya estamos en una transacción, usarla directamente
	if ds.tx != nil {
		log.Println("[UOW][Atomic()] Ya hay una transacción activa, usando la transacción existente")
		return cb(ds)
	}

	// Iniciar una transacción
	log.Println("[UOW][Atomic()] Iniciando nueva transacción")
	tx, err := ds.pool.Begin(ctx)
	if err != nil {
		log.Println("[UOW][Atomic()] Error al iniciar transacción:", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer para manejar commit o rollback
	defer func() {
		if p := recover(); p != nil {
			// En caso de panic, hacer rollback
			log.Println("[UOW][Atomic()] Panic detectado, haciendo rollback:", p)
			tx.Rollback(ctx)
			panic(p) // Re-panic después del rollback
		} else if err != nil {
			// En caso de error, hacer rollback
			log.Println("[UOW][Atomic()] Error detectado, haciendo rollback:", err)
			rbErr := tx.Rollback(ctx)
			if rbErr != nil {
				log.Println("[UOW][Atomic()] Error al hacer rollback:", rbErr)
				err = fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
			}
		} else {
			// Si todo está bien, hacer commit
			log.Println("[UOW][Atomic()] Transacción exitosa, haciendo commit")
			err = tx.Commit(ctx)
			if err != nil {
				log.Println("[UOW][Atomic()] Error al hacer commit:", err)
				err = fmt.Errorf("failed to commit transaction: %w", err)
			} else {
				log.Println("[UOW][Atomic()] Commit exitoso")
			}
		}
	}()

	// Crear una nueva instancia de DataStore con la transacción
	dataStoreTx := ds.WithTx(tx)

	// Ejecutar el callback
	log.Println("[UOW][Atomic()] Ejecutando callback dentro de la transacción")
	err = cb(dataStoreTx)
	if err != nil {
		log.Println("[UOW][Atomic()] Error en callback:", err)
	}

	return err
}

func (ds *DataStore) Job() ports.IJobRepository {
	return NewJobRepository(ds.tx, ds.pool)
}
func (ds *DataStore) Attempt() ports.IAttemptRepository {
	return NewAttemptRepository(ds.tx, ds.pool)
}
func (ds *DataStore) Event() ports.IEventRepository {
	return NewEventRepository(ds.tx, ds.pool)
}

// func (ds *DataStore) DeadLetter() ports.IDeadLetterRepository {
// 	return NewDeadLetterRepository(ds.tx, ds.pool)
// }
