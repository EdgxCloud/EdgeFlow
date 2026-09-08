/**
 * Node catalog
 *
 * Caches the backend node-type catalog so the canvas can tell how many output
 * ports a node has. Routing nodes (if, switch, filter) deliver a message only to
 * the port their rules selected, so each port needs its own connection handle --
 * otherwise every edge is created on port 0 and the other branches are dead.
 *
 * The catalog is fetched once per session and shared by every caller.
 */

import { useEffect, useState } from 'react'
import { nodesApi } from './nodes'

type OutputNames = string[]

let catalog: Map<string, OutputNames> | null = null
let inFlight: Promise<Map<string, OutputNames>> | null = null

/** Loads the catalog, reusing an in-flight or completed request. */
export function loadNodeCatalog(): Promise<Map<string, OutputNames>> {
  if (catalog) return Promise.resolve(catalog)

  if (!inFlight) {
    inFlight = nodesApi
      .getTypes()
      .then((response) => {
        const loaded = new Map<string, OutputNames>()
        for (const nodeType of response.node_types ?? []) {
          const key = nodeType.type || nodeType.id
          if (!key) continue
          loaded.set(
            key,
            (nodeType.outputs ?? []).map((output, index) => output.name || `${index}`)
          )
        }
        catalog = loaded
        return loaded
      })
      .catch(() => {
        // Leave the cache empty and allow a later retry; callers fall back to a
        // single output, which is correct for the great majority of nodes.
        inFlight = null
        return new Map<string, OutputNames>()
      })
  }

  return inFlight
}

/** Output port names for a node type, or undefined if not loaded yet. */
export function getOutputNames(nodeType: string): OutputNames | undefined {
  return catalog?.get(nodeType)
}

/**
 * Output port names for a node type. Returns a single unnamed output until the
 * catalog resolves, then re-renders with the real ports.
 */
export function useOutputPorts(nodeType: string): OutputNames {
  const [ports, setPorts] = useState<OutputNames>(() => getOutputNames(nodeType) ?? [''])

  useEffect(() => {
    let active = true
    loadNodeCatalog().then((loaded) => {
      if (!active) return
      const found = loaded.get(nodeType)
      setPorts(found && found.length > 0 ? found : [''])
    })
    return () => {
      active = false
    }
  }, [nodeType])

  return ports
}
