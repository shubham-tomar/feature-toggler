# feature-toggler
Toggling Features in a Live App (Without Deployment)

## Table of Contents
- [Overview](#overview)
- [Features](#features)


## Overview

Feature toggling is a technique used to control the availability of new features in a software application. It allows developers to enable or disable features dynamically, without the need to deploy new code or restart the application. This is particularly useful in scenarios where you want to test new features in a production environment or roll out features gradually to a subset of users.

## Testing via GraphQl Playground
- Open the GraphQl Playground at http://localhost:8080/graphql

- Variables
{
  "key": "feature-key",
  "flagId": "feature-flag-id-here",
  "projectId": "project-id-here"
}

- Use the following query to get current user:
```graphql
query Me {
  me {
    id
    name
    email
  }
}
```

- Use the following query to create project:
```graphql
mutation CreateProject {
  createProject(name: "Test Project") {
    id
    name
    members {
      id
      role
      user {
        id
        name
        email
      }
    }
    created_at
    updated_at
  }
}
```

- Use the following query to get all projects:
```graphql
query GetProjects {
  projects {
    id
    name
    members {
      id
      role
      user {
        id
        name
      }
    }
    created_at
    updated_at
  }
}
```

- Use the following query to get project by id:
```graphql
query GetProject($projectId: ID!) {
  project(id: $projectId) {
    id
    name
    members {
      id
      role
      user {
        id
        name
        email
      }
    }
    createdAt
    updatedAt
  }
}
```

- Use the following query to create feature flag:
```graphql
mutation CreateFeatureFlag {
  createFeatureFlag(input: {
    projectId: "project-id-here",
    key: "new-feature",
    name: "New Feature",
    description: "This is a new feature flag",
    initialStates: [
      { environment: DEV, enabled: true },
      { environment: STAGING, enabled: false },
      { environment: PROD, enabled: false }
    ]
  }) {
    id
    key
    name
    description
    createdAt
    updatedAt
    states {
      id
      environment
      enabled
      updatedBy {
        id
        name
      }
      updatedAt
    }
    project {
      id
      name
    }
    created_by {
      id
      name
    }
  }
}
```

- Use the following query to toggle feature flag:
```graphql
mutation ToggleFeature {
  toggleFeatureFlag(input: {
    featureFlagId: "feature-flag-id-here",
    environment: PROD,
    enabled: true
  }) {
    id
    environment
    enabled
    updatedAt
    updated_by {
      id
      name
    }
    feature_flag {
      id
      key
      name
    }
  }
}
```

- Use the following query to get feature flag by id:
```graphql
query GetFeatureFlag($flagId: ID!) {
  feature_flag(id: $flagId) {
    id
    key
    name
    description
    states {
      id
      environment
      enabled
      updatedAt
      updated_by {
        id
        name
      }
    }
    project {
      id
      name
    }
    created_by {
      id
      name
    }
    createdAt
    updatedAt
  }
}
```

- Use the following query to get feature flag by key:
```graphql
query GetFeatureFlagByKey($key: String!) {
  feature_flag_by_key(key: $key) {
    id
    key
    name
    description
    states {
      id
      environment
      enabled
    }
  }
}
```




