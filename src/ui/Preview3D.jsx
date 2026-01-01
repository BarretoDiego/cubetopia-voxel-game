/**
 * 3D Preview Component for Gallery and Inventory
 * Uses @react-three/drei View for efficient shared-context rendering
 */

import React, { useRef, Suspense } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { View, Stage, PerspectiveCamera, Environment, Bounds } from '@react-three/drei';
import * as THREE from 'three';

// --- Models ---

function BlockModel({ color }) {
  return (
    <mesh rotation={[0.5, 0.5, 0]}>
      <boxGeometry args={[1, 1, 1]} />
      <meshStandardMaterial 
        color={color || '#808080'} 
        roughness={0.6}
        metalness={0.1}
        envMapIntensity={1}
      />
    </mesh>
  );
}

function CreatureModel({ type, color }) {
  const group = useRef();
  
  const materialProps = {
    color: color || '#22CC22',
    roughness: 0.6,
    metalness: 0.1,
    envMapIntensity: 0.8
  };

  const renderBody = () => {
    switch(type) {
      case 'slime':
        return (
          <group>
             {/* Body */}
            <mesh position={[0, 0.4, 0]}>
              <boxGeometry args={[1, 0.8, 1]} />
              <meshPhysicalMaterial {...materialProps} transparent opacity={0.8} />
            </mesh>
            {/* Eyes */}
            <mesh position={[-0.2, 0.5, 0.52]}>
              <boxGeometry args={[0.15, 0.15, 0.05]} />
              <meshBasicMaterial color="white" />
            </mesh>
            <mesh position={[0.2, 0.5, 0.52]}>
              <boxGeometry args={[0.15, 0.15, 0.05]} />
              <meshBasicMaterial color="white" />
            </mesh>
            <mesh position={[-0.2, 0.5, 0.55]}>
              <boxGeometry args={[0.08, 0.08, 0.05]} />
              <meshBasicMaterial color="black" />
            </mesh>
            <mesh position={[0.2, 0.5, 0.55]}>
              <boxGeometry args={[0.08, 0.08, 0.05]} />
              <meshBasicMaterial color="black" />
            </mesh>
          </group>
        );
      case 'pig':
        return (
          <group>
            <mesh position={[0, 0.4, 0]}>
              <boxGeometry args={[0.8, 0.5, 1.2]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
            <mesh position={[0, 0.5, 0.7]}>
              <boxGeometry args={[0.5, 0.5, 0.5]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
            <mesh position={[0, 0.4, 0.96]}>
              <boxGeometry args={[0.25, 0.2, 0.1]} />
              <meshStandardMaterial color="#FF69B4" roughness={0.5} />
            </mesh>
             {/* Legs */}
             {[-0.25, 0.25].map(x => (
               [-0.35, 0.35].map(z => (
                 <mesh key={`${x}-${z}`} position={[x, 0.15, z]}>
                    <boxGeometry args={[0.15, 0.3, 0.15]} />
                    <meshStandardMaterial {...materialProps} />
                 </mesh>
               ))
             ))}
          </group>
        );
      case 'zombie':
        return (
          <group position={[0, -0.5, 0]}> {/* Adjust center */}
            {/* Body */}
            <mesh position={[0, 0.8, 0]}>
              <boxGeometry args={[0.5, 0.7, 0.3]} />
              <meshStandardMaterial color="#4169E1" roughness={0.8} />
            </mesh>
            {/* Head */}
            <mesh position={[0, 1.35, 0]}>
              <boxGeometry args={[0.4, 0.4, 0.4]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
            {/* Arms */}
            <mesh position={[-0.35, 0.9, 0.2]} rotation={[-Math.PI / 3, 0, 0]}>
              <boxGeometry args={[0.15, 0.6, 0.15]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
            <mesh position={[0.35, 0.9, 0.2]} rotation={[-Math.PI / 3, 0, 0]}>
              <boxGeometry args={[0.15, 0.6, 0.15]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
            {/* Legs */}
             <mesh position={[-0.15, 0.3, 0]}>
              <boxGeometry args={[0.2, 0.6, 0.2]} />
              <meshStandardMaterial color="#483D8B" />
            </mesh>
            <mesh position={[0.15, 0.3, 0]}>
              <boxGeometry args={[0.2, 0.6, 0.2]} />
              <meshStandardMaterial color="#483D8B" />
            </mesh>
          </group>
        );
        case 'spider':
          return (
             <group position={[0, 0.2, 0]}>
                <mesh position={[0, 0.3, -0.3]}>
                  <boxGeometry args={[0.6, 0.4, 0.8]} />
                  <meshStandardMaterial {...materialProps} />
                </mesh>
                <mesh position={[0, 0.25, 0.2]}>
                  <boxGeometry args={[0.4, 0.3, 0.4]} />
                  <meshStandardMaterial {...materialProps} />
                </mesh>
                {/* Legs */}
                {[0, 1, 2, 3].map(i => (
                  <group key={i}>
                    <mesh position={[-0.4, 0.15, -0.2 + i * 0.15]} rotation={[0, 0, -0.3]}>
                      <boxGeometry args={[0.5, 0.05, 0.05]} />
                      <meshStandardMaterial {...materialProps} />
                    </mesh>
                    <mesh position={[0.4, 0.15, -0.2 + i * 0.15]} rotation={[0, 0, 0.3]}>
                      <boxGeometry args={[0.5, 0.05, 0.05]} />
                      <meshStandardMaterial {...materialProps} />
                    </mesh>
                  </group>
                ))}
             </group>
          );
      case 'cow':
        return (
          <group>
            <mesh position={[0, 0.5, 0]}>
              <boxGeometry args={[0.8, 0.6, 1.2]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
             {/* Spots */}
            <mesh position={[0.2, 0.6, 0.61]}>
               <boxGeometry args={[0.25, 0.25, 0.02]} />
               <meshStandardMaterial color="#1A1A1A" />
            </mesh>
             <mesh position={[-0.15, 0.45, 0.61]}>
               <boxGeometry args={[0.2, 0.2, 0.02]} />
               <meshStandardMaterial color="#1A1A1A" />
            </mesh>
            {/* Head */}
            <mesh position={[0, 0.7, 0.75]}>
              <boxGeometry args={[0.5, 0.5, 0.4]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
             {/* Horns */}
            <mesh position={[-0.2, 1, 0.7]}>
               <boxGeometry args={[0.08, 0.2, 0.08]} />
               <meshStandardMaterial color="#F5DEB3" />
            </mesh>
            <mesh position={[0.2, 1, 0.7]}>
               <boxGeometry args={[0.08, 0.2, 0.08]} />
               <meshStandardMaterial color="#F5DEB3" />
            </mesh>
             {/* Legs */}
             {[-0.25, 0.25].map(x => (
               [-0.4, 0.4].map(z => (
                 <mesh key={`${x}-${z}`} position={[x, 0.2, z]}>
                    <boxGeometry args={[0.15, 0.4, 0.15]} />
                    <meshStandardMaterial {...materialProps} />
                 </mesh>
               ))
             ))}
          </group>
        );
      case 'bird':
         return (
           <group position={[0, 0.3, 0]}>
             <mesh>
               <boxGeometry args={[0.3, 0.3, 0.5]} />
               <meshStandardMaterial {...materialProps} />
             </mesh>
             <mesh position={[0, 0.4, 0.3]}>
               <boxGeometry args={[0.2, 0.2, 0.25]} />
               <meshStandardMaterial {...materialProps} />
             </mesh>
             <mesh position={[0, 0.35, 0.47]}>
               <boxGeometry args={[0.1, 0.08, 0.15]} />
               <meshStandardMaterial color="#FFA500" />
             </mesh>
             {/* Wings */}
             <mesh position={[-0.3, 0.35, 0]}>
               <boxGeometry args={[0.4, 0.05, 0.2]} />
               <meshStandardMaterial {...materialProps} />
             </mesh>
              <mesh position={[0.3, 0.35, 0]}>
               <boxGeometry args={[0.4, 0.05, 0.2]} />
               <meshStandardMaterial {...materialProps} />
             </mesh>
           </group>
         );
      default: // Cube fallback
         return (
          <mesh position={[0, 0.4, 0]}>
            <boxGeometry args={[0.8, 0.8, 0.8]} />
            <meshStandardMaterial {...materialProps} />
          </mesh>
         );
    }
  }

  return (
    <group ref={group}>
      {renderBody()}
    </group>
  );
}

function ItemModel({ type, color }) {
  const materialProps = {
    color: color || '#C0C0C0',
    metalness: 0.8,
    roughness: 0.2,
    envMapIntensity: 1.2
  };
  const handleMaterial = <meshStandardMaterial color="#8B4513" roughness={0.9} />;
  
  const renderItem = () => {
    switch (type) {
      case 'sword':
        return (
          <group rotation={[0, 0, -Math.PI / 4]}>
            <mesh position={[0, -0.5, 0]}>
               <boxGeometry args={[0.1, 0.4, 0.1]} />
               {handleMaterial}
            </mesh>
            <mesh position={[0, -0.25, 0]}>
               <boxGeometry args={[0.4, 0.1, 0.1]} />
               <meshStandardMaterial {...materialProps} />
            </mesh>
            <mesh position={[0, 0.3, 0]}>
               <boxGeometry args={[0.12, 1, 0.05]} />
               <meshStandardMaterial {...materialProps} />
            </mesh>
            <mesh position={[0, 0.9, 0]}>
               <coneGeometry args={[0.06, 0.2, 4]} />
               <meshStandardMaterial {...materialProps} />
            </mesh>
          </group>
        );
      case 'pickaxe':
        return (
          <group rotation={[0, 0, -Math.PI / 4]} position={[0, 0.2, 0]}>
            <mesh>
               <boxGeometry args={[0.1, 1.2, 0.1]} />
               {handleMaterial}
            </mesh>
            <mesh position={[0, 0.5, 0]}>
               <boxGeometry args={[0.8, 0.15, 0.15]} />
               <meshStandardMaterial {...materialProps} />
            </mesh>
            <mesh position={[-0.5, 0.6, 0]} rotation={[0, 0, Math.PI / 2]}>
               <coneGeometry args={[0.08, 0.25, 4]} />
               <meshStandardMaterial {...materialProps} />
            </mesh>
             <mesh position={[0.5, 0.6, 0]} rotation={[0, 0, -Math.PI / 2]}>
               <coneGeometry args={[0.08, 0.25, 4]} />
               <meshStandardMaterial {...materialProps} />
            </mesh>
          </group>
        );
      case 'axe':
         return (
           <group rotation={[0, 0, -Math.PI / 4]}>
             <mesh>
               <boxGeometry args={[0.1, 1.2, 0.1]} />
               {handleMaterial}
             </mesh>
             <mesh position={[0.2, 0.4, 0]}>
               <boxGeometry args={[0.5, 0.6, 0.1]} />
               <meshStandardMaterial {...materialProps} />
             </mesh>
           </group>
         );
      case 'diamond':
        return (
           <group>
             <mesh position={[0, 0.15, 0]}>
               <coneGeometry args={[0.4, 0.5, 8]} />
               <meshPhysicalMaterial 
                  color="#00FFFF"
                  metalness={0.1}
                  roughness={0}
                  transmission={0.6}
                  thickness={0.5}
                  envMapIntensity={2}
               />
             </mesh>
             <mesh position={[0, -0.15, 0]} rotation={[Math.PI, 0, 0]}>
               <coneGeometry args={[0.4, 0.5, 8]} />
               <meshPhysicalMaterial 
                  color="#00FFFF"
                  metalness={0.1}
                  roughness={0}
                  transmission={0.6}
                  thickness={0.5}
                  envMapIntensity={2}
               />
             </mesh>
           </group>
        );
      case 'apple':
        return (
          <group>
             <mesh>
               <sphereGeometry args={[0.4, 16, 16]} />
               <meshStandardMaterial color="red" roughness={0.3} />
             </mesh>
             <mesh position={[0, 0.45, 0]}>
                <cylinderGeometry args={[0.03, 0.03, 0.15]} />
                 {handleMaterial}
             </mesh>
              <mesh position={[0.1, 0.5, 0]} rotation={[0, 0, 0.3]}>
                <boxGeometry args={[0.15, 0.05, 0.1]} />
                <meshStandardMaterial color="green" />
             </mesh>
          </group>
        );
      default:
         return (
            <mesh>
              <boxGeometry args={[0.5, 0.5, 0.5]} />
              <meshStandardMaterial {...materialProps} />
            </mesh>
         );
    }
  }

  return (
    <group>
      {renderItem()}
    </group>
  );
}


// --- Main Components ---

export const PreviewCanvas = ({ children }) => {
  return (
    <div className="preview-canvas-container" style={{
      position: 'fixed',
      top: 0,
      left: 0,
      width: '100vw',
      height: '100vh',
      pointerEvents: 'none', // Let clicks pass through to UI
      zIndex: 9999, // On top of everything
    }}>
      <Canvas
        eventSource={document.getElementById('root')}
        style={{ pointerEvents: 'none' }}
        gl={{ alpha: true, antialias: true }}
      >
        <View.Port />
        {children}
      </Canvas>
    </div>
  );
};

// Simple auto-rotation component
function AutoRotate() {
  const ref = useRef();
  
  useFrame((state, delta) => {
    if (ref.current) {
      ref.current.rotation.y += delta * 0.5;
    }
  });

  return <group ref={ref} />;
}


export const Preview3D = ({ type, variant, color, size = 100 }) => {
  const containerRef = useRef(null);
  
  // Decide what to render based on type (block, creature, item)
  const renderContent = () => {
    if (type === 'block') return <BlockModel color={color} />;
    if (type === 'creature') return <CreatureModel type={variant} color={color} />;
    if (type === 'item') return <ItemModel type={variant} color={color} />;
    return <BlockModel color={color} />;
  };

  return (
    <div 
      ref={containerRef} 
      className="preview-3d-container"
      style={{ 
        width: size, 
        height: size, 
        display: 'inline-block',
        position: 'relative'
      }}
    >
      <View track={containerRef}>
        <PerspectiveCamera makeDefault position={[2, 2, 2]} fov={50} />
        <Environment preset="sunset" />
        <ambientLight intensity={0.5} />
        <pointLight position={[10, 10, 5]} intensity={1} />
        
        <Suspense fallback={null}>
            <Bounds fit clip observe margin={1.2}>
               <Stage 
                  intensity={0.5} 
                  environment="city" 
                  adjustCamera={false}
                  shadows={{ type: 'contact', opacity: 0.5, blur: 3 }}
               >
                 {renderContent()}
               </Stage>
            </Bounds>
        </Suspense>
        
        {/* Rotation Animation - Parent to the content to rotate it */}
        <group>
           <AutoRotate /> 
           {/* We need to rotate the content, so we wrap rendering in a group that AutoRotate affects? 
               Actually AutoRotate just rotates its own ref group. 
               Let's make AutoRotate wrap children.
           */}
        </group>
      </View>
    </div>
  );
};

// Re-export old names for compatibility if needed, but we should update usage
export const BlockPreview3D = ({ color, size }) => (
  <Preview3D type="block" color={color} size={size} />
);

export const CreaturePreview3D = ({ type, color, size }) => (
  <Preview3D type="creature" variant={type} color={color} size={size} />
);

export const ItemPreview3D = ({ type, color, size }) => (
  <Preview3D type="item" variant={type} color={color} size={size} />
);

export default { PreviewCanvas, Preview3D, BlockPreview3D, CreaturePreview3D, ItemPreview3D };
