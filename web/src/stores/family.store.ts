import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface Child {
  id: string
  name: string
  dateOfBirth: string
  gender?: string
  avatarUrl?: string
}

export interface Family {
  id: string
  name: string
  children: Child[]
}

interface FamilyState {
  currentFamily: Family | null
  currentChild: Child | null
  families: Family[]
  setCurrentFamily: (family: Family) => void
  setCurrentChild: (child: Child) => void
  setFamilies: (families: Family[]) => void
  addChild: (child: Child) => void
  updateChild: (child: Child) => void
  removeChild: (childId: string) => void
}

export const useFamilyStore = create<FamilyState>()(
  persist(
    (set, get) => ({
      currentFamily: null,
      currentChild: null,
      families: [],

      setCurrentFamily: (family) =>
        set({
          currentFamily: family,
          currentChild: family.children[0] || null,
        }),

      setCurrentChild: (child) =>
        set({
          currentChild: child,
        }),

      setFamilies: (families) =>
        set({
          families,
          currentFamily: families[0] || null,
          currentChild: families[0]?.children[0] || null,
        }),

      addChild: (child) => {
        const family = get().currentFamily
        if (!family) return

        const updatedFamily = {
          ...family,
          children: [...family.children, child],
        }

        set({
          currentFamily: updatedFamily,
          families: get().families.map((f) =>
            f.id === family.id ? updatedFamily : f
          ),
        })
      },

      updateChild: (child) => {
        const family = get().currentFamily
        if (!family) return

        const updatedFamily = {
          ...family,
          children: family.children.map((c) =>
            c.id === child.id ? child : c
          ),
        }

        set({
          currentFamily: updatedFamily,
          currentChild:
            get().currentChild?.id === child.id ? child : get().currentChild,
          families: get().families.map((f) =>
            f.id === family.id ? updatedFamily : f
          ),
        })
      },

      removeChild: (childId) => {
        const family = get().currentFamily
        if (!family) return

        const updatedFamily = {
          ...family,
          children: family.children.filter((c) => c.id !== childId),
        }

        set({
          currentFamily: updatedFamily,
          currentChild:
            get().currentChild?.id === childId
              ? updatedFamily.children[0] || null
              : get().currentChild,
          families: get().families.map((f) =>
            f.id === family.id ? updatedFamily : f
          ),
        })
      },
    }),
    {
      name: 'family-storage',
      partialize: (state) => ({
        currentFamily: state.currentFamily,
        currentChild: state.currentChild,
        families: state.families,
      }),
    }
  )
)
