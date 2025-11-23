package cloud

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/com_818cloud/shode/pkg/registry"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config describes the cloud registry connections.
type Config struct {
	DatabaseURL string

	S3Endpoint  string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool
	S3Region    string
}

// Service coordinates PostgreSQL metadata + S3 storage.
type Service struct {
	db     *sql.DB
	s3     *minio.Client
	bucket string
	region string
}

// NewService wires DB + S3 clients.
func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init s3: %w", err)
	}

	if err := ensureBucket(ctx, client, cfg.S3Bucket, cfg.S3Region); err != nil {
		return nil, err
	}

	svc := &Service{
		db:     db,
		s3:     client,
		bucket: cfg.S3Bucket,
		region: cfg.S3Region,
	}

	if err := svc.bootstrap(ctx); err != nil {
		return nil, err
	}

	return svc, nil
}

func ensureBucket(ctx context.Context, client *minio.Client, bucket, region string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}
	return nil
}

func (s *Service) bootstrap(ctx context.Context) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS packages (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			author TEXT,
			license TEXT,
			homepage TEXT,
			repository TEXT,
			keywords JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			downloads BIGINT NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS package_versions (
			id SERIAL PRIMARY KEY,
			package_id INTEGER NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
			version TEXT NOT NULL,
			description TEXT,
			author TEXT,
			dependencies JSONB,
			dev_dependencies JSONB,
			checksum TEXT,
			signature TEXT,
			signature_algo TEXT,
			signer_id TEXT,
			tarball_key TEXT NOT NULL,
			published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE(package_id, version)
		);`,
	}

	for _, stmt := range schema {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("bootstrap schema: %w", err)
		}
	}
	return nil
}

// Publish stores metadata in Postgres and uploads tarball to S3.
func (s *Service) Publish(ctx context.Context, req *registry.PublishRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	packageID, err := s.upsertPackage(ctx, tx, req.Package)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s.tar.gz", req.Package.Name, req.Package.Version)
	if err := s.uploadTarball(ctx, key, req.Tarball); err != nil {
		return err
	}

	if err := s.upsertVersion(ctx, tx, packageID, req, key); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) upsertPackage(
	ctx context.Context,
	tx *sql.Tx,
	pkg *registry.Package,
) (int64, error) {
	keywords, _ := json.Marshal(pkg.Keywords)
	query := `INSERT INTO packages
		(name, description, author, license, homepage, repository, keywords, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,now(),now())
		ON CONFLICT (name) DO UPDATE
		SET description=excluded.description,
			author=excluded.author,
			license=excluded.license,
			homepage=excluded.homepage,
			repository=excluded.repository,
			keywords=excluded.keywords,
			updated_at=now()
		RETURNING id`

	var id int64
	if err := tx.QueryRowContext(
		ctx,
		query,
		pkg.Name,
		pkg.Description,
		pkg.Author,
		pkg.License,
		pkg.Homepage,
		pkg.Repository,
		keywords,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) upsertVersion(
	ctx context.Context,
	tx *sql.Tx,
	packageID int64,
	req *registry.PublishRequest,
	tarballKey string,
) error {
	deps, _ := json.Marshal(req.Package.Dependencies)
	devDeps, _ := json.Marshal(req.Package.DevDependencies)

	query := `INSERT INTO package_versions
		(package_id, version, description, author, dependencies, dev_dependencies, checksum, signature, signature_algo, signer_id, tarball_key, published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())
		ON CONFLICT (package_id, version) DO UPDATE
		SET description=excluded.description,
			author=excluded.author,
			dependencies=excluded.dependencies,
			dev_dependencies=excluded.dev_dependencies,
			checksum=excluded.checksum,
			signature=excluded.signature,
			signature_algo=excluded.signature_algo,
			signer_id=excluded.signer_id,
			tarball_key=excluded.tarball_key,
			published_at=excluded.published_at`

	_, err := tx.ExecContext(
		ctx,
		query,
		packageID,
		req.Package.Version,
		req.Package.Description,
		req.Package.Author,
		deps,
		devDeps,
		req.Checksum,
		req.Signature,
		req.SignatureAlgo,
		req.SignerID,
		tarballKey,
	)
	return err
}

func (s *Service) uploadTarball(ctx context.Context, key string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := s.s3.PutObject(ctx, s.bucket, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/gzip",
	})
	return err
}

// GetPackage retrieves package metadata with versions.
func (s *Service) GetPackage(ctx context.Context, name string) (*registry.PackageMetadata, error) {
	var meta registry.PackageMetadata
	err := s.db.QueryRowContext(ctx, `SELECT id, description, author, license, homepage, repository, keywords, created_at, updated_at, downloads
		FROM packages WHERE name=$1`, name).
		Scan(
			new(int64),
			&meta.Description,
			&meta.Author,
			&meta.License,
			&meta.Homepage,
			&meta.Repository,
		&jsonColumn{&meta.Keywords},
			&meta.CreatedAt,
			&meta.UpdatedAt,
			&meta.Downloads,
		)
	if err != nil {
		return nil, err
	}

	meta.Name = name
	meta.Versions = make(map[string]*registry.PackageVersion)

	rows, err := s.db.QueryContext(ctx, `SELECT version, description, author, checksum, signature, signature_algo, signer_id, tarball_key, published_at
		FROM package_versions pv
		INNER JOIN packages p ON pv.package_id = p.id
		WHERE p.name=$1
		ORDER BY published_at DESC`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version registry.PackageVersion
		var tarballKey string
		if err := rows.Scan(
			&version.Version,
			&version.Description,
			&version.Author,
			&version.Shasum,
			&version.Signature,
			&version.SignatureAlgo,
			&version.SignerID,
			&tarballKey,
			&version.PublishedAt,
		); err != nil {
			return nil, err
		}
		version.TarballURL = fmt.Sprintf("s3://%s/%s", s.bucket, tarballKey)
		meta.Versions[version.Version] = &version
		if meta.LatestVersion == "" {
			meta.LatestVersion = version.Version
		}
	}

	return &meta, nil
}

// SearchPackages performs simple ILIKE matches.
func (s *Service) SearchPackages(ctx context.Context, query *registry.SearchQuery) ([]*registry.SearchResult, error) {
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}
	pattern := fmt.Sprintf("%%%s%%", query.Query)
	rows, err := s.db.QueryContext(ctx, `SELECT name, description, author, downloads
		FROM packages
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY downloads DESC
		LIMIT $2 OFFSET $3`, pattern, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*registry.SearchResult
	for rows.Next() {
		var res registry.SearchResult
		if err := rows.Scan(&res.Name, &res.Description, &res.Author, &res.Downloads); err != nil {
			return nil, err
		}
		results = append(results, &res)
	}
	return results, nil
}

// GetDownloadURL returns a presigned URL for the version tarball.
func (s *Service) GetDownloadURL(ctx context.Context, name, version string, expiry time.Duration) (string, error) {
	var tarballKey string
	err := s.db.QueryRowContext(ctx, `SELECT tarball_key
		FROM package_versions pv
		INNER JOIN packages p ON pv.package_id = p.id
		WHERE p.name=$1 AND pv.version=$2`, name, version).
		Scan(&tarballKey)
	if err != nil {
		return "", err
	}

	url, err := s.s3.PresignedGetObject(ctx, s.bucket, tarballKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

type jsonColumn struct {
	target interface{}
}

func (c *jsonColumn) Scan(src interface{}) error {
	if c.target == nil {
		return nil
	}

	if src == nil {
		switch t := c.target.(type) {
		case *[]string:
			*t = nil
		default:
		}
		return nil
	}

	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported json column type %T", src)
	}

	return json.Unmarshal(data, c.target)
}
