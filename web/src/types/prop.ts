export interface Prop {
    id: number
    drama_id: number
    name: string
    type?: string
    description?: string
    prompt?: string
    image_url?: string
    reference_images?: any
    image_orientation?: string
    created_at: string
    updated_at: string
}

export interface CreatePropRequest {
    drama_id: number
    name: string
    type?: string
    description?: string
    prompt?: string
    image_url?: string
}

export interface UpdatePropRequest {
    name?: string
    type?: string
    description?: string
    prompt?: string
    image_url?: string
    reference_images?: string[]
    image_orientation?: string
}
