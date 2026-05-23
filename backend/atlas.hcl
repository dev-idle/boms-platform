// Atlas project config — declarative schema in db/schema.hcl, versioned migrations in migrations/.
// Requires POSTGRES_URL for apply; dev database uses Docker Postgres (see Makefile ATLAS_DEV_URL).
//
// Install CLI: curl -sSf https://atlasgo.sh | sh
// Windows: download from https://release.ariga.io/atlas/atlas-windows-amd64-latest.exe

variable "database_url" {
  type    = string
  default = getenv("POSTGRES_URL")
}

variable "dev_url" {
  type    = string
  default = "docker://postgres/16/dev?search_path=public"
}

env "local" {
  src = "file://db/schema.hcl"
  dev = var.dev_url
  url = var.database_url

  migration {
    dir = "file://migrations"
    # Store revision history in public (avoids atlas_schema_revisions schema bootstrap issues on Neon).
    revisions_schema = "public"
  }

  diff {
    concurrent_index {
      add  = true
      drop = true
    }
  }

  lint {
    destructive {
      error = true
    }
  }
}
