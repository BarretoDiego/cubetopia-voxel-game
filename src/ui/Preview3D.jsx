/**
 * 3D Preview Component for Gallery
 * Renders small 3D voxel models for items, blocks, and creatures
 */

import React, { useRef, useEffect } from 'react';
import * as THREE from 'three';

/**
 * Renders a 3D preview of a block
 */
export function BlockPreview3D({ color, size = 60 }) {
  const containerRef = useRef(null);
  const rendererRef = useRef(null);
  
  useEffect(() => {
    if (!containerRef.current) return;
    
    // Scene setup
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 100);
    camera.position.set(2, 2, 2);
    camera.lookAt(0, 0, 0);
    
    // Renderer
    const renderer = new THREE.WebGLRenderer({ 
      alpha: true, 
      antialias: true,
      preserveDrawingBuffer: true 
    });
    renderer.setSize(size, size);
    renderer.setClearColor(0x000000, 0);
    containerRef.current.appendChild(renderer.domElement);
    rendererRef.current = renderer;
    
    // Lighting
    const ambient = new THREE.AmbientLight(0xffffff, 0.6);
    scene.add(ambient);
    const directional = new THREE.DirectionalLight(0xffffff, 0.8);
    directional.position.set(2, 3, 1);
    scene.add(directional);
    
    // Block cube
    const geometry = new THREE.BoxGeometry(1, 1, 1);
    const material = new THREE.MeshStandardMaterial({ 
      color: color || '#808080',
      roughness: 0.7,
      metalness: 0.1
    });
    const cube = new THREE.Mesh(geometry, material);
    scene.add(cube);
    
    // Animate rotation
    let animationId;
    const animate = () => {
      animationId = requestAnimationFrame(animate);
      cube.rotation.y += 0.02;
      renderer.render(scene, camera);
    };
    animate();
    
    // Cleanup
    return () => {
      cancelAnimationFrame(animationId);
      if (containerRef.current && renderer.domElement) {
        containerRef.current.removeChild(renderer.domElement);
      }
      geometry.dispose();
      material.dispose();
      renderer.dispose();
    };
  }, [color, size]);
  
  return <div ref={containerRef} style={{ width: size, height: size }} />;
}

/**
 * Renders a 3D preview of a creature
 */
export function CreaturePreview3D({ type, color, size = 60 }) {
  const containerRef = useRef(null);
  const rendererRef = useRef(null);
  
  useEffect(() => {
    if (!containerRef.current) return;
    
    // Scene setup
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 100);
    camera.position.set(2.5, 2, 2.5);
    camera.lookAt(0, 0.5, 0);
    
    // Renderer
    const renderer = new THREE.WebGLRenderer({ 
      alpha: true, 
      antialias: true 
    });
    renderer.setSize(size, size);
    renderer.setClearColor(0x000000, 0);
    containerRef.current.appendChild(renderer.domElement);
    rendererRef.current = renderer;
    
    // Lighting
    const ambient = new THREE.AmbientLight(0xffffff, 0.6);
    scene.add(ambient);
    const directional = new THREE.DirectionalLight(0xffffff, 0.8);
    directional.position.set(2, 3, 1);
    scene.add(directional);
    
    // Create creature based on type
    const group = new THREE.Group();
    const material = new THREE.MeshStandardMaterial({ 
      color: color || '#22CC22',
      roughness: 0.6
    });
    
    if (type === 'slime') {
      // Slime - rounded cube
      const body = new THREE.Mesh(new THREE.BoxGeometry(1, 0.8, 1), material);
      body.position.y = 0.4;
      group.add(body);
      // Eyes
      const eyeMat = new THREE.MeshBasicMaterial({ color: 0xffffff });
      const eyeL = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.15, 0.05), eyeMat);
      eyeL.position.set(-0.2, 0.5, 0.52);
      const eyeR = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.15, 0.05), eyeMat);
      eyeR.position.set(0.2, 0.5, 0.52);
      group.add(eyeL, eyeR);
      // Pupils
      const pupilMat = new THREE.MeshBasicMaterial({ color: 0x000000 });
      const pupilL = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.08, 0.05), pupilMat);
      pupilL.position.set(-0.2, 0.5, 0.55);
      const pupilR = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.08, 0.05), pupilMat);
      pupilR.position.set(0.2, 0.5, 0.55);
      group.add(pupilL, pupilR);
    } else if (type === 'pig') {
      // Pig body
      const body = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.5, 1.2), material);
      body.position.y = 0.4;
      group.add(body);
      // Head
      const head = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.5, 0.5), material);
      head.position.set(0, 0.5, 0.7);
      group.add(head);
      // Snout
      const snoutMat = new THREE.MeshStandardMaterial({ color: 0xFF69B4 });
      const snout = new THREE.Mesh(new THREE.BoxGeometry(0.25, 0.2, 0.1), snoutMat);
      snout.position.set(0, 0.4, 0.96);
      group.add(snout);
      // Legs
      for (let x of [-0.25, 0.25]) {
        for (let z of [-0.35, 0.35]) {
          const leg = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.3, 0.15), material);
          leg.position.set(x, 0.15, z);
          group.add(leg);
        }
      }
    } else if (type === 'zombie') {
      // Zombie body
      const bodyMat = new THREE.MeshStandardMaterial({ color: 0x4169E1 });
      const body = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.7, 0.3), bodyMat);
      body.position.y = 0.8;
      group.add(body);
      // Head
      const head = new THREE.Mesh(new THREE.BoxGeometry(0.4, 0.4, 0.4), material);
      head.position.y = 1.35;
      group.add(head);
      // Arms
      const armL = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.6, 0.15), material);
      armL.position.set(-0.35, 0.9, 0.2);
      armL.rotation.x = -Math.PI / 3;
      const armR = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.6, 0.15), material);
      armR.position.set(0.35, 0.9, 0.2);
      armR.rotation.x = -Math.PI / 3;
      group.add(armL, armR);
      // Legs
      const legMat = new THREE.MeshStandardMaterial({ color: 0x483D8B });
      const legL = new THREE.Mesh(new THREE.BoxGeometry(0.2, 0.6, 0.2), legMat);
      legL.position.set(-0.15, 0.3, 0);
      const legR = new THREE.Mesh(new THREE.BoxGeometry(0.2, 0.6, 0.2), legMat);
      legR.position.set(0.15, 0.3, 0);
      group.add(legL, legR);
    } else if (type === 'spider') {
      // Spider body
      const abdomen = new THREE.Mesh(new THREE.BoxGeometry(0.6, 0.4, 0.8), material);
      abdomen.position.set(0, 0.3, -0.3);
      group.add(abdomen);
      const thorax = new THREE.Mesh(new THREE.BoxGeometry(0.4, 0.3, 0.4), material);
      thorax.position.set(0, 0.25, 0.2);
      group.add(thorax);
      // Eyes
      const eyeMat = new THREE.MeshBasicMaterial({ color: 0xFF0000 });
      for (let i = 0; i < 4; i++) {
        const eye = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.08, 0.08), eyeMat);
        eye.position.set(-0.12 + i * 0.08, 0.35, 0.42);
        group.add(eye);
      }
      // Legs
      for (let i = 0; i < 4; i++) {
        const legL = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.05, 0.05), material);
        legL.position.set(-0.4, 0.15, -0.2 + i * 0.15);
        legL.rotation.z = -0.3;
        const legR = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.05, 0.05), material);
        legR.position.set(0.4, 0.15, -0.2 + i * 0.15);
        legR.rotation.z = 0.3;
        group.add(legL, legR);
      }
    } else if (type === 'bird') {
      // Bird body
      const body = new THREE.Mesh(new THREE.BoxGeometry(0.3, 0.3, 0.5), material);
      body.position.y = 0.3;
      group.add(body);
      // Head
      const head = new THREE.Mesh(new THREE.BoxGeometry(0.2, 0.2, 0.25), material);
      head.position.set(0, 0.4, 0.3);
      group.add(head);
      // Beak
      const beakMat = new THREE.MeshStandardMaterial({ color: 0xFFA500 });
      const beak = new THREE.Mesh(new THREE.BoxGeometry(0.1, 0.08, 0.15), beakMat);
      beak.position.set(0, 0.35, 0.47);
      group.add(beak);
      // Wings
      const wingL = new THREE.Mesh(new THREE.BoxGeometry(0.4, 0.05, 0.2), material);
      wingL.position.set(-0.3, 0.35, 0);
      const wingR = new THREE.Mesh(new THREE.BoxGeometry(0.4, 0.05, 0.2), material);
      wingR.position.set(0.3, 0.35, 0);
      group.add(wingL, wingR);
    } else if (type === 'cow') {
      // Cow body
      const body = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.6, 1.2), material);
      body.position.y = 0.5;
      group.add(body);
      // Spots
      const spotMat = new THREE.MeshStandardMaterial({ color: 0x1A1A1A });
      const spot1 = new THREE.Mesh(new THREE.BoxGeometry(0.25, 0.25, 0.02), spotMat);
      spot1.position.set(0.2, 0.6, 0.61);
      const spot2 = new THREE.Mesh(new THREE.BoxGeometry(0.2, 0.2, 0.02), spotMat);
      spot2.position.set(-0.15, 0.45, 0.61);
      group.add(spot1, spot2);
      // Head
      const head = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.5, 0.4), material);
      head.position.set(0, 0.7, 0.75);
      group.add(head);
      // Horns
      const hornMat = new THREE.MeshStandardMaterial({ color: 0xF5DEB3 });
      const hornL = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.2, 0.08), hornMat);
      hornL.position.set(-0.2, 1, 0.7);
      const hornR = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.2, 0.08), hornMat);
      hornR.position.set(0.2, 1, 0.7);
      group.add(hornL, hornR);
      // Legs
      for (let x of [-0.25, 0.25]) {
        for (let z of [-0.4, 0.4]) {
          const leg = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.4, 0.15), material);
          leg.position.set(x, 0.2, z);
          group.add(leg);
        }
      }
    } else {
      // Default cube
      const cube = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.8, 0.8), material);
      cube.position.y = 0.4;
      group.add(cube);
    }
    
    scene.add(group);
    
    // Animate rotation
    let animationId;
    const animate = () => {
      animationId = requestAnimationFrame(animate);
      group.rotation.y += 0.02;
      renderer.render(scene, camera);
    };
    animate();
    
    // Cleanup
    return () => {
      cancelAnimationFrame(animationId);
      if (containerRef.current && renderer.domElement) {
        containerRef.current.removeChild(renderer.domElement);
      }
      renderer.dispose();
    };
  }, [type, color, size]);
  
  return <div ref={containerRef} style={{ width: size, height: size }} />;
}

/**
 * Renders a 3D preview of an item (weapon/tool)
 */
export function ItemPreview3D({ type, color, size = 60 }) {
  const containerRef = useRef(null);
  
  useEffect(() => {
    if (!containerRef.current) return;
    
    // Scene setup
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 100);
    camera.position.set(0, 0, 3);
    camera.lookAt(0, 0, 0);
    
    // Renderer
    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
    renderer.setSize(size, size);
    renderer.setClearColor(0x000000, 0);
    containerRef.current.appendChild(renderer.domElement);
    
    // Lighting
    const ambient = new THREE.AmbientLight(0xffffff, 0.7);
    scene.add(ambient);
    const directional = new THREE.DirectionalLight(0xffffff, 0.6);
    directional.position.set(1, 2, 2);
    scene.add(directional);
    
    // Create item
    const group = new THREE.Group();
    const handleMat = new THREE.MeshStandardMaterial({ color: 0x8B4513 });
    const bladeMat = new THREE.MeshStandardMaterial({ color: color || 0xC0C0C0, metalness: 0.8, roughness: 0.2 });
    
    if (type === 'sword') {
      // Handle
      const handle = new THREE.Mesh(new THREE.BoxGeometry(0.1, 0.4, 0.1), handleMat);
      handle.position.y = -0.5;
      group.add(handle);
      // Guard
      const guard = new THREE.Mesh(new THREE.BoxGeometry(0.4, 0.1, 0.1), bladeMat);
      guard.position.y = -0.25;
      group.add(guard);
      // Blade
      const blade = new THREE.Mesh(new THREE.BoxGeometry(0.12, 1, 0.05), bladeMat);
      blade.position.y = 0.3;
      group.add(blade);
      // Tip
      const tip = new THREE.Mesh(new THREE.ConeGeometry(0.06, 0.2, 4), bladeMat);
      tip.position.y = 0.9;
      group.add(tip);
    } else if (type === 'pickaxe') {
      // Handle
      const handle = new THREE.Mesh(new THREE.BoxGeometry(0.1, 1.2, 0.1), handleMat);
      group.add(handle);
      // Head
      const head = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.15, 0.15), bladeMat);
      head.position.y = 0.5;
      group.add(head);
      // Pick points
      const pointL = new THREE.Mesh(new THREE.ConeGeometry(0.08, 0.25, 4), bladeMat);
      pointL.position.set(-0.5, 0.6, 0);
      pointL.rotation.z = Math.PI / 2;
      const pointR = new THREE.Mesh(new THREE.ConeGeometry(0.08, 0.25, 4), bladeMat);
      pointR.position.set(0.5, 0.6, 0);
      pointR.rotation.z = -Math.PI / 2;
      group.add(pointL, pointR);
    } else if (type === 'axe') {
      // Handle
      const handle = new THREE.Mesh(new THREE.BoxGeometry(0.1, 1.2, 0.1), handleMat);
      group.add(handle);
      // Axe head
      const head = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.6, 0.1), bladeMat);
      head.position.set(0.2, 0.4, 0);
      group.add(head);
    } else if (type === 'apple') {
      // Apple body
      const appleMat = new THREE.MeshStandardMaterial({ color: 0xFF0000 });
      const body = new THREE.Mesh(new THREE.SphereGeometry(0.4, 16, 16), appleMat);
      group.add(body);
      // Stem
      const stem = new THREE.Mesh(new THREE.CylinderGeometry(0.03, 0.03, 0.15), handleMat);
      stem.position.y = 0.45;
      group.add(stem);
      // Leaf
      const leafMat = new THREE.MeshStandardMaterial({ color: 0x228B22 });
      const leaf = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.05, 0.1), leafMat);
      leaf.position.set(0.1, 0.5, 0);
      leaf.rotation.z = 0.3;
      group.add(leaf);
    } else if (type === 'diamond') {
      // Diamond shape
      const diamondMat = new THREE.MeshStandardMaterial({ 
        color: 0x00FFFF, 
        metalness: 0.9, 
        roughness: 0.1,
        transparent: true,
        opacity: 0.9
      });
      const topGeo = new THREE.ConeGeometry(0.4, 0.5, 8);
      const top = new THREE.Mesh(topGeo, diamondMat);
      top.position.y = 0.15;
      group.add(top);
      const bottomGeo = new THREE.ConeGeometry(0.4, 0.5, 8);
      const bottom = new THREE.Mesh(bottomGeo, diamondMat);
      bottom.position.y = -0.15;
      bottom.rotation.x = Math.PI;
      group.add(bottom);
    } else {
      // Default cube
      const cube = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.5, 0.5), bladeMat);
      group.add(cube);
    }
    
    group.rotation.z = -0.3;
    scene.add(group);
    
    // Animate rotation
    let animationId;
    const animate = () => {
      animationId = requestAnimationFrame(animate);
      group.rotation.y += 0.02;
      renderer.render(scene, camera);
    };
    animate();
    
    // Cleanup
    return () => {
      cancelAnimationFrame(animationId);
      if (containerRef.current && renderer.domElement) {
        containerRef.current.removeChild(renderer.domElement);
      }
      renderer.dispose();
    };
  }, [type, color, size]);
  
  return <div ref={containerRef} style={{ width: size, height: size }} />;
}

export default { BlockPreview3D, CreaturePreview3D, ItemPreview3D };
