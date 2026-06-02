export interface Application {
  id: string
  name: string
  repoUrl: string
  branch: string
  composePath: string
  composeContent?: string
  envVars: string
  status: string
  source: string
  createdAt: string
  updatedAt: string
}

export interface AppDomain {
  id: string
  applicationId: string
  host: string
  port: number
  path: string
  internalPath: string
  stripPath: boolean
  https: boolean
  serviceName: string
  createdAt: string
  updatedAt: string
}

export interface Deployment {
  id: string
  applicationId: string
  status: string
  commitSha: string
  commitMessage: string
  logPath: string
  createdAt: string
  updatedAt: string
}

export interface DashboardItem extends Application {
  latestDeployment: Deployment | null
}
