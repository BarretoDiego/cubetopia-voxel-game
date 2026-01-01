
import React, { useRef, useMemo, useEffect } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

const PARTICLE_COUNT = 1000;
const RANGE = 40;

export function EnvironmentParticles({ playerPosition, disabled }) {
  const meshRef = useRef();
  
  // Create dummy object for instance positioning
  const dummy = useMemo(() => new THREE.Object3D(), []);
  
  // Initial positions
  const particles = useMemo(() => {
    const temp = [];
    for (let i = 0; i < PARTICLE_COUNT; i++) {
      temp.push({
        x: (Math.random() - 0.5) * RANGE,
        y: (Math.random() - 0.5) * RANGE,
        z: (Math.random() - 0.5) * RANGE,
        speed: 0.2 + Math.random() * 0.5,
        offset: Math.random() * Math.PI * 2
      });
    }
    return temp;
  }, []);

  useFrame((state) => {
    if (!meshRef.current || disabled) return;

    const time = state.clock.getElapsedTime();
    const px = playerPosition?.x || 0;
    const py = playerPosition?.y || 0;
    const pz = playerPosition?.z || 0;

    particles.forEach((particle, i) => {
      // Float movement
      let x = particle.x + Math.sin(time * 0.1 + particle.offset) * 2;
      let y = particle.y + Math.sin(time * 0.3 + particle.offset) * 1.5;
      let z = particle.z + Math.cos(time * 0.15 + particle.offset) * 2;

      // Wrap around player locally
      // We calculate world pos relative to player to create "infinite" field
      const wx = (x + px) % RANGE;
      const wy = (y + py + 10) % RANGE; // +10 offset for height bias
      const wz = (z + pz) % RANGE;

      // Adjust to be centered on player visual range
      dummy.position.set(
        px + (wx - px + RANGE * 1.5) % RANGE - RANGE / 2,
        py + (wy - py + RANGE * 1.5) % RANGE - RANGE / 2,
        pz + (wz - pz + RANGE * 1.5) % RANGE - RANGE / 2
      );

      // Subtle rotation
      dummy.rotation.x = time * 0.1 + particle.offset;
      dummy.rotation.y = time * 0.2 + particle.offset;
      
      // Pulsing scale
      const scale = 0.5 + Math.sin(time * 2 + particle.offset) * 0.2;
      dummy.scale.set(scale, scale, scale);

      dummy.updateMatrix();
      meshRef.current.setMatrixAt(i, dummy.matrix);
    });

    meshRef.current.instanceMatrix.needsUpdate = true;
  });

  if (disabled) return null;

  return (
    <instancedMesh ref={meshRef} args={[null, null, PARTICLE_COUNT]}>
      <dodecahedronGeometry args={[0.05, 0]} />
      <meshBasicMaterial 
        color="#ffffff" 
        transparent 
        opacity={0.4} 
        blending={THREE.AdditiveBlending} 
      />
    </instancedMesh>
  );
}
