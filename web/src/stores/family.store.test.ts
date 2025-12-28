import { describe, it, expect, beforeEach } from 'vitest'
import { useFamilyStore, type Child, type Family } from './family.store'

describe('useFamilyStore', () => {
  const mockChild1: Child = {
    id: 'child-1',
    name: 'Emma',
    dateOfBirth: '2023-01-15',
    gender: 'female',
  }

  const mockChild2: Child = {
    id: 'child-2',
    name: 'Liam',
    dateOfBirth: '2024-03-20',
    gender: 'male',
  }

  const mockFamily: Family = {
    id: 'family-1',
    name: 'Test Family',
    children: [mockChild1, mockChild2],
  }

  const mockFamily2: Family = {
    id: 'family-2',
    name: 'Second Family',
    children: [],
  }

  beforeEach(() => {
    // Reset store to initial state before each test
    useFamilyStore.setState({
      currentFamily: null,
      currentChild: null,
      families: [],
    })
  })

  describe('initial state', () => {
    it('should start with no family selected', () => {
      const state = useFamilyStore.getState()
      expect(state.currentFamily).toBeNull()
      expect(state.currentChild).toBeNull()
      expect(state.families).toEqual([])
    })
  })

  describe('setCurrentFamily', () => {
    it('should set current family and first child', () => {
      useFamilyStore.getState().setCurrentFamily(mockFamily)

      const state = useFamilyStore.getState()
      expect(state.currentFamily).toEqual(mockFamily)
      expect(state.currentChild).toEqual(mockChild1)
    })

    it('should set current child to null if family has no children', () => {
      useFamilyStore.getState().setCurrentFamily(mockFamily2)

      const state = useFamilyStore.getState()
      expect(state.currentFamily).toEqual(mockFamily2)
      expect(state.currentChild).toBeNull()
    })
  })

  describe('setCurrentChild', () => {
    it('should set current child', () => {
      useFamilyStore.getState().setCurrentFamily(mockFamily)
      useFamilyStore.getState().setCurrentChild(mockChild2)

      expect(useFamilyStore.getState().currentChild).toEqual(mockChild2)
    })
  })

  describe('setFamilies', () => {
    it('should set families and select first family/child', () => {
      useFamilyStore.getState().setFamilies([mockFamily, mockFamily2])

      const state = useFamilyStore.getState()
      expect(state.families).toHaveLength(2)
      expect(state.currentFamily).toEqual(mockFamily)
      expect(state.currentChild).toEqual(mockChild1)
    })

    it('should handle empty families array', () => {
      useFamilyStore.getState().setFamilies([])

      const state = useFamilyStore.getState()
      expect(state.families).toEqual([])
      expect(state.currentFamily).toBeNull()
      expect(state.currentChild).toBeNull()
    })
  })

  describe('addChild', () => {
    it('should add child to current family', () => {
      useFamilyStore.getState().setFamilies([mockFamily])

      const newChild: Child = {
        id: 'child-3',
        name: 'Noah',
        dateOfBirth: '2025-01-01',
      }
      useFamilyStore.getState().addChild(newChild)

      const state = useFamilyStore.getState()
      expect(state.currentFamily?.children).toHaveLength(3)
      expect(state.currentFamily?.children[2]).toEqual(newChild)
    })

    it('should update only the current family in families array', () => {
      useFamilyStore.getState().setFamilies([mockFamily, mockFamily2])

      const newChild: Child = {
        id: 'child-3',
        name: 'Noah',
        dateOfBirth: '2025-01-01',
      }
      useFamilyStore.getState().addChild(newChild)

      const state = useFamilyStore.getState()
      // Current family should have new child
      expect(state.families[0].children).toHaveLength(3)
      // Other family should be unchanged
      expect(state.families[1].children).toHaveLength(0)
    })

    it('should not add child if no current family', () => {
      const newChild: Child = {
        id: 'child-3',
        name: 'Noah',
        dateOfBirth: '2025-01-01',
      }
      useFamilyStore.getState().addChild(newChild)

      expect(useFamilyStore.getState().currentFamily).toBeNull()
    })
  })

  describe('updateChild', () => {
    it('should update only the current family in families array', () => {
      useFamilyStore.getState().setFamilies([mockFamily, mockFamily2])

      const updatedChild: Child = {
        ...mockChild1,
        name: 'Updated Emma',
      }
      useFamilyStore.getState().updateChild(updatedChild)

      const state = useFamilyStore.getState()
      // Current family should have updated child
      expect(state.families[0].children[0].name).toBe('Updated Emma')
      // Other family should be unchanged
      expect(state.families[1]).toEqual(mockFamily2)
    })

    it('should update child in current family', () => {
      useFamilyStore.getState().setFamilies([mockFamily])

      const updatedChild: Child = {
        ...mockChild1,
        name: 'Emma Updated',
      }
      useFamilyStore.getState().updateChild(updatedChild)

      const state = useFamilyStore.getState()
      expect(state.currentFamily?.children[0].name).toBe('Emma Updated')
    })

    it('should update current child if it matches', () => {
      useFamilyStore.getState().setFamilies([mockFamily])
      expect(useFamilyStore.getState().currentChild?.id).toBe('child-1')

      const updatedChild: Child = {
        ...mockChild1,
        name: 'Emma Updated',
      }
      useFamilyStore.getState().updateChild(updatedChild)

      expect(useFamilyStore.getState().currentChild?.name).toBe('Emma Updated')
    })

    it('should not update current child if different child updated', () => {
      useFamilyStore.getState().setFamilies([mockFamily])

      const updatedChild: Child = {
        ...mockChild2,
        name: 'Liam Updated',
      }
      useFamilyStore.getState().updateChild(updatedChild)

      expect(useFamilyStore.getState().currentChild?.id).toBe('child-1')
      expect(useFamilyStore.getState().currentChild?.name).toBe('Emma')
    })

    it('should not update if no current family', () => {
      useFamilyStore.getState().updateChild(mockChild1)
      expect(useFamilyStore.getState().currentFamily).toBeNull()
    })
  })

  describe('removeChild', () => {
    it('should update only the current family in families array', () => {
      useFamilyStore.getState().setFamilies([mockFamily, mockFamily2])
      useFamilyStore.getState().removeChild('child-2')

      const state = useFamilyStore.getState()
      // Current family should have child removed
      expect(state.families[0].children).toHaveLength(1)
      // Other family should be unchanged
      expect(state.families[1]).toEqual(mockFamily2)
    })

    it('should remove child from current family', () => {
      useFamilyStore.getState().setFamilies([mockFamily])
      useFamilyStore.getState().removeChild('child-2')

      const state = useFamilyStore.getState()
      expect(state.currentFamily?.children).toHaveLength(1)
      expect(state.currentFamily?.children[0].id).toBe('child-1')
    })

    it('should select next child if current child removed', () => {
      useFamilyStore.getState().setFamilies([mockFamily])
      expect(useFamilyStore.getState().currentChild?.id).toBe('child-1')

      useFamilyStore.getState().removeChild('child-1')

      expect(useFamilyStore.getState().currentChild?.id).toBe('child-2')
    })

    it('should set current child to null if last child removed', () => {
      const singleChildFamily: Family = {
        id: 'family-1',
        name: 'Single Child Family',
        children: [mockChild1],
      }
      useFamilyStore.getState().setFamilies([singleChildFamily])

      useFamilyStore.getState().removeChild('child-1')

      expect(useFamilyStore.getState().currentChild).toBeNull()
    })

    it('should not remove if no current family', () => {
      useFamilyStore.getState().removeChild('child-1')
      expect(useFamilyStore.getState().currentFamily).toBeNull()
    })
  })
})
