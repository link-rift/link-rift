package worker

import (
	"net"

	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

// GeoLookup provides IP-to-location resolution using a MaxMind GeoIP2 database.
type GeoLookup struct {
	reader *geoip2.Reader
	logger *zap.Logger
}

// NewGeoLookup opens the MaxMind .mmdb database at the given path.
// Returns nil, nil if path is empty (opt-out).
func NewGeoLookup(dbPath string, logger *zap.Logger) (*GeoLookup, error) {
	if dbPath == "" {
		return nil, nil
	}

	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	logger.Info("GeoIP2 database loaded", zap.String("path", dbPath))
	return &GeoLookup{reader: reader, logger: logger}, nil
}

// Lookup resolves an IP address to country, region, and city.
// Returns empty strings on failure (best-effort).
func (g *GeoLookup) Lookup(ipStr string) (country, region, city string) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", "", ""
	}

	record, err := g.reader.City(ip)
	if err != nil {
		g.logger.Debug("GeoIP lookup failed", zap.String("ip", ipStr), zap.Error(err))
		return "", "", ""
	}

	country = record.Country.IsoCode
	if len(record.Subdivisions) > 0 {
		region = record.Subdivisions[0].Names["en"]
	}
	city = record.City.Names["en"]

	return country, region, city
}

// Close releases the GeoIP2 database resources.
func (g *GeoLookup) Close() error {
	if g.reader != nil {
		return g.reader.Close()
	}
	return nil
}
