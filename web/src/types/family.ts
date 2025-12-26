export interface Child {
  id: string
  name: string
  dateOfBirth: string
  gender?: 'male' | 'female' | 'other'
  avatarUrl?: string
  createdAt: string
  updatedAt: string
}

export interface Family {
  id: string
  name: string
  children: Child[]
  createdAt: string
  updatedAt: string
}

export interface CreateFamilyRequest {
  name: string
}

export interface CreateChildRequest {
  name: string
  date_of_birth: string
  gender?: string
}
