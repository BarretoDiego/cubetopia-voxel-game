/**
 * Hook para controles do jogador
 */

import { useRef, useEffect, useCallback } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";
import {
  GRAVITY,
  JUMP_FORCE,
  PLAYER_SPEED,
  SPRINT_MULTIPLIER,
  CHUNK_HEIGHT,
} from "../utils/constants.js";
import { BlockTypes } from "../core/blocks/BlockTypes.js";
import { blockRegistry } from "../core/blocks/BlockRegistry.js";

export function usePlayerControls(
  world,
  onPositionChange,
  initialPosition = null
) {
  const { camera } = useThree();
  const initialized = useRef(false);

  // Referências de estado
  const position = useRef(new THREE.Vector3(0, 50, 0));
  const velocity = useRef(new THREE.Vector3(0, 0, 0));
  const isOnGround = useRef(false);

  // Set initial position once when provided
  if (initialPosition && !initialized.current) {
    position.current.set(
      initialPosition.x,
      initialPosition.y,
      initialPosition.z
    );
    camera.position.copy(position.current);
    initialized.current = true;
  }

  // Pre-allocated vectors to avoid GC pressure
  const moveDirection = useRef(new THREE.Vector3());
  const cameraDirection = useRef(new THREE.Vector3());
  const newPos = useRef(new THREE.Vector3());
  const yAxis = useRef(new THREE.Vector3(0, 1, 0));

  // Teclas pressionadas
  const keys = useRef({
    forward: false,
    backward: false,
    left: false,
    right: false,
    jump: false,
    sprint: false,
  });

  // Event handlers
  useEffect(() => {
    const onKeyDown = (e) => {
      switch (e.code) {
        case "KeyW":
          keys.current.forward = true;
          break;
        case "KeyS":
          keys.current.backward = true;
          break;
        case "KeyA":
          keys.current.right = true;
          break;
        case "KeyD":
          keys.current.left = true;
          break;
        case "Space":
          keys.current.jump = true;
          e.preventDefault();
          break;
        case "ShiftLeft":
          keys.current.sprint = true;
          break;
        case "KeyG":
          // Deglitch: teleport player 10 units up
          position.current.y += 10;
          velocity.current.set(0, 0, 0);
          isOnGround.current = false;
          break;
      }
    };

    const onKeyUp = (e) => {
      switch (e.code) {
        case "KeyW":
          keys.current.forward = false;
          break;
        case "KeyS":
          keys.current.backward = false;
          break;
        case "KeyA":
          keys.current.right = false;
          break;
        case "KeyD":
          keys.current.left = false;
          break;
        case "Space":
          keys.current.jump = false;
          break;
        case "ShiftLeft":
          keys.current.sprint = false;
          break;
      }
    };

    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("keyup", onKeyUp);

    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("keyup", onKeyUp);
    };
  }, []);

  // Game loop
  useFrame((state, delta) => {
    if (!world) return;

    // Limita delta para evitar bugs em frames longos
    const dt = Math.min(delta, 0.1);

    // Calcula direção do movimento (reusing pre-allocated vector)
    const moveDir = moveDirection.current;
    moveDir.set(0, 0, 0);

    // Movement direction relative to camera
    // Note: PointerLockControls uses +Z as forward
    if (keys.current.forward) moveDir.z += 1; // W = forward (+Z)
    if (keys.current.backward) moveDir.z -= 1; // S = backward (-Z)
    if (keys.current.left) moveDir.x -= 1; // A = left
    if (keys.current.right) moveDir.x += 1; // D = right

    // Normaliza e aplica rotação da câmera
    if (moveDir.lengthSq() > 0) {
      moveDir.normalize();

      // Rotaciona direção baseada na câmera (apenas Y)
      camera.getWorldDirection(cameraDirection.current);
      const angle = Math.atan2(
        cameraDirection.current.x,
        cameraDirection.current.z
      );

      moveDir.applyAxisAngle(yAxis.current, angle);
    }

    // Velocidade
    const speed = PLAYER_SPEED * (keys.current.sprint ? SPRINT_MULTIPLIER : 1);

    // Atualiza velocidade horizontal
    velocity.current.x = moveDir.x * speed;
    velocity.current.z = moveDir.z * speed;

    // Gravidade
    velocity.current.y -= GRAVITY * dt;

    // Pulo
    if (keys.current.jump && isOnGround.current) {
      velocity.current.y = JUMP_FORCE;
      isOnGround.current = false;
    }

    // Movimento com colisão (reusing pre-allocated vector)
    const pos = newPos.current;
    pos.copy(position.current);

    // Move X
    pos.x += velocity.current.x * dt;
    if (checkCollision(world, pos.x, position.current.y, position.current.z)) {
      pos.x = position.current.x;
      velocity.current.x = 0;
    }

    // Move Z
    pos.z += velocity.current.z * dt;
    if (checkCollision(world, pos.x, position.current.y, pos.z)) {
      pos.z = position.current.z;
      velocity.current.z = 0;
    }

    // Move Y
    pos.y += velocity.current.y * dt;

    // Verifica colisão com chão
    if (velocity.current.y < 0) {
      if (checkCollision(world, pos.x, pos.y - 1.6, pos.z)) {
        pos.y = Math.floor(pos.y - 0.6) + 1.6;
        velocity.current.y = 0;
        isOnGround.current = true;
      }
    }

    // Verifica colisão com teto
    if (velocity.current.y > 0) {
      if (checkCollision(world, pos.x, pos.y + 0.2, pos.z)) {
        velocity.current.y = 0;
      }
    }

    // Limita altura
    pos.y = Math.max(2, Math.min(CHUNK_HEIGHT - 2, pos.y));

    // Atualiza posição
    position.current.copy(pos);
    camera.position.copy(position.current);

    // Callback de posição
    if (onPositionChange) {
      onPositionChange({
        x: Math.floor(position.current.x),
        y: Math.floor(position.current.y),
        z: Math.floor(position.current.z),
      });
    }
  });

  return {
    position: position.current,
    velocity: velocity.current,
    isOnGround: isOnGround.current,
  };
}

/**
 * Verifica colisão em uma posição
 */
function checkCollision(world, x, y, z) {
  // Verifica múltiplos pontos ao redor do jogador
  const checkPoints = [
    [0, 0, 0],
    [-0.3, 0, -0.3],
    [0.3, 0, -0.3],
    [-0.3, 0, 0.3],
    [0.3, 0, 0.3],
  ];

  for (const [dx, dy, dz] of checkPoints) {
    const block = world.getBlock(
      Math.floor(x + dx),
      Math.floor(y + dy),
      Math.floor(z + dz)
    );

    if (blockRegistry.isCollidable(block)) {
      return true;
    }
  }

  return false;
}

export default usePlayerControls;
