/**
 * Player 3D Model - Voxel-style player character
 * Visible in 3rd person view
 */

import React, { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

/**
 * Creates a voxel-style player model (Minecraft-like)
 */
export function PlayerModel({ position, rotation, isMoving, isThirdPerson }) {
  const groupRef = useRef();
  const leftArmRef = useRef();
  const rightArmRef = useRef();
  const leftLegRef = useRef();
  const rightLegRef = useRef();
  const walkCycleRef = useRef(0);

  // Only render in 3rd person
  if (!isThirdPerson) return null;

  // Materials
  const skinMaterial = useMemo(() => 
    new THREE.MeshStandardMaterial({ color: '#e8beac', roughness: 0.7 }), []
  );
  const shirtMaterial = useMemo(() => 
    new THREE.MeshStandardMaterial({ color: '#42a5f5', roughness: 0.8 }), []
  );
  const pantsMaterial = useMemo(() => 
    new THREE.MeshStandardMaterial({ color: '#1565c0', roughness: 0.8 }), []
  );
  const shoeMaterial = useMemo(() => 
    new THREE.MeshStandardMaterial({ color: '#37474f', roughness: 0.9 }), []
  );
  const hairMaterial = useMemo(() => 
    new THREE.MeshStandardMaterial({ color: '#5d4037', roughness: 0.9 }), []
  );

  // Walk animation
  useFrame((state, delta) => {
    if (!groupRef.current) return;

    // Update position and rotation
    groupRef.current.position.set(position.x, position.y - 0.5, position.z);
    groupRef.current.rotation.y = -rotation + Math.PI; // Face away from camera

    // Walk cycle animation
    if (isMoving) {
      walkCycleRef.current += delta * 8;
      const swing = Math.sin(walkCycleRef.current) * 0.5;
      
      if (leftArmRef.current) leftArmRef.current.rotation.x = swing;
      if (rightArmRef.current) rightArmRef.current.rotation.x = -swing;
      if (leftLegRef.current) leftLegRef.current.rotation.x = -swing;
      if (rightLegRef.current) rightLegRef.current.rotation.x = swing;
    } else {
      walkCycleRef.current = 0;
      if (leftArmRef.current) leftArmRef.current.rotation.x = 0;
      if (rightArmRef.current) rightArmRef.current.rotation.x = 0;
      if (leftLegRef.current) leftLegRef.current.rotation.x = 0;
      if (rightLegRef.current) rightLegRef.current.rotation.x = 0;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Head */}
      <mesh position={[0, 1.5, 0]} material={skinMaterial}>
        <boxGeometry args={[0.5, 0.5, 0.5]} />
      </mesh>
      
      {/* Hair */}
      <mesh position={[0, 1.8, 0]} material={hairMaterial}>
        <boxGeometry args={[0.52, 0.15, 0.52]} />
      </mesh>
      
      {/* Body/Shirt */}
      <mesh position={[0, 0.9, 0]} material={shirtMaterial}>
        <boxGeometry args={[0.5, 0.7, 0.3]} />
      </mesh>
      
      {/* Left Arm */}
      <group position={[-0.35, 1.05, 0]} ref={leftArmRef}>
        <mesh position={[0, -0.25, 0]} material={shirtMaterial}>
          <boxGeometry args={[0.2, 0.35, 0.2]} />
        </mesh>
        <mesh position={[0, -0.55, 0]} material={skinMaterial}>
          <boxGeometry args={[0.18, 0.25, 0.18]} />
        </mesh>
      </group>
      
      {/* Right Arm */}
      <group position={[0.35, 1.05, 0]} ref={rightArmRef}>
        <mesh position={[0, -0.25, 0]} material={shirtMaterial}>
          <boxGeometry args={[0.2, 0.35, 0.2]} />
        </mesh>
        <mesh position={[0, -0.55, 0]} material={skinMaterial}>
          <boxGeometry args={[0.18, 0.25, 0.18]} />
        </mesh>
      </group>
      
      {/* Left Leg */}
      <group position={[-0.12, 0.4, 0]} ref={leftLegRef}>
        <mesh position={[0, -0.15, 0]} material={pantsMaterial}>
          <boxGeometry args={[0.22, 0.4, 0.22]} />
        </mesh>
        <mesh position={[0, -0.45, 0]} material={shoeMaterial}>
          <boxGeometry args={[0.22, 0.15, 0.28]} />
        </mesh>
      </group>
      
      {/* Right Leg */}
      <group position={[0.12, 0.4, 0]} ref={rightLegRef}>
        <mesh position={[0, -0.15, 0]} material={pantsMaterial}>
          <boxGeometry args={[0.22, 0.4, 0.22]} />
        </mesh>
        <mesh position={[0, -0.45, 0]} material={shoeMaterial}>
          <boxGeometry args={[0.22, 0.15, 0.28]} />
        </mesh>
      </group>
    </group>
  );
}

export default PlayerModel;
