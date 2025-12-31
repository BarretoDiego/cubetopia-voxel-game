/**
 * Hook para interação com blocos
 */

import { useCallback, useRef } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { BlockTypes } from "../core/blocks/BlockTypes.js";
import { blockRegistry } from "../core/blocks/BlockRegistry.js";

export function useBlockInteraction(world, selectedBlockType, onWorldChange) {
  const { camera } = useThree();
  const lastClickTime = useRef(0);
  const CLICK_COOLDOWN = 150; // ms

  /**
   * Raycast para encontrar bloco alvo
   */
  const raycast = useCallback(() => {
    if (!world) return null;

    const direction = new THREE.Vector3();
    camera.getWorldDirection(direction);

    const origin = camera.position.clone();
    const step = 0.1;
    const maxDistance = 6;

    let prev = null;
    let current = origin.clone();

    for (let d = 0; d < maxDistance; d += step) {
      prev = current.clone();
      current.addScaledVector(direction, step);

      const bx = Math.floor(current.x);
      const by = Math.floor(current.y);
      const bz = Math.floor(current.z);

      const blockType = world.getBlock(bx, by, bz);

      if (blockType !== BlockTypes.AIR && !blockRegistry.isLiquid(blockType)) {
        return {
          hit: { x: bx, y: by, z: bz },
          prev: {
            x: Math.floor(prev.x),
            y: Math.floor(prev.y),
            z: Math.floor(prev.z),
          },
          blockType,
          distance: d,
        };
      }
    }

    return null;
  }, [camera, world]);

  /**
   * Quebra bloco
   */
  const breakBlock = useCallback(() => {
    const now = Date.now();
    if (now - lastClickTime.current < CLICK_COOLDOWN) return false;
    lastClickTime.current = now;

    const result = raycast();
    if (!result) return false;

    const { hit, blockType } = result;

    // Verifica se bloco é quebrável
    const blockDef = blockRegistry.get(blockType);
    if (blockDef.indestructible) return false;

    // Remove bloco
    world.setBlock(hit.x, hit.y, hit.z, BlockTypes.AIR);

    if (onWorldChange) {
      onWorldChange();
    }

    return true;
  }, [raycast, world, onWorldChange]);

  /**
   * Coloca bloco
   */
  const placeBlock = useCallback(() => {
    const now = Date.now();
    if (now - lastClickTime.current < CLICK_COOLDOWN) return false;
    lastClickTime.current = now;

    const result = raycast();
    if (!result) return false;

    const { prev } = result;

    // Verifica se posição é válida (não está dentro do jogador)
    const playerPos = camera.position;
    const dx = Math.abs(prev.x - playerPos.x);
    const dy = Math.abs(prev.y - playerPos.y);
    const dz = Math.abs(prev.z - playerPos.z);

    if (dx < 0.8 && dy < 1.8 && dz < 0.8) {
      return false; // Muito perto do jogador
    }

    // Coloca bloco
    world.setBlock(prev.x, prev.y, prev.z, selectedBlockType);

    if (onWorldChange) {
      onWorldChange();
    }

    return true;
  }, [raycast, world, selectedBlockType, camera, onWorldChange]);

  /**
   * Obtém bloco sendo olhado
   */
  const getTargetBlock = useCallback(() => {
    return raycast();
  }, [raycast]);

  return {
    breakBlock,
    placeBlock,
    getTargetBlock,
  };
}

export default useBlockInteraction;
