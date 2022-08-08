# Infrapi - Frontend for infrared

utility for [lhridder/infrared](https://github.com/lhridder/infrared) / [lhridder/infrapi](https://github.com/lhridder/infrapi)

### Global config.yml
```yaml
bind: :5000
debug: false
redis:
  host: localhost
  pass:
  db: 0
```

## Used sources
- [Redis library for golang](https://github.com/go-redis/redis/v8)
- [Chi router](https://github.com/go-chi/chi)