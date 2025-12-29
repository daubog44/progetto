#!/bin/bash

# Docker "Garbage Collection" Script
# Goal: Clean up unused Docker resources (dangling images, build cache, stopped containers)
# WITHOUT deleting the resources currently used by the running project.

echo "============================================"
echo "Docker Space Reclamation Script"
echo "============================================"

# Check if project is running
if [ -z "$(docker compose ps -q)" ]; then
    echo "WARNING: The project appears to be STOPPED."
    echo "Running cleanup now might remove images you want to keep if you use 'docker image prune -a'."
    echo "To protect your project images, please start the project first: 'docker compose up -d'"
    echo ""
    read -p "Do you want to continue anyway? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborting."
        exit 1
    fi
else
    echo "Project is RUNNING. Active containers and their images are safe."
fi

echo ""
echo "Step 1: Pruning dangling images (untagged, <none>)"
# These are intermediate layers from builds that are no longer used. Safe to remove.
docker image prune -f

echo ""
echo "Step 2: Pruning build cache"
# This frees up MASSIVE space but means next build will be slower (no cache hit).
docker builder prune -f

echo ""
echo "Step 3: Pruning stopped containers"
# This removes containers that are not running. 
# WARNING: If you stopped your project, this deletes the containers (but not volumes).
docker container prune -f

echo ""
echo "Step 4: Pruning unused networks"
docker network prune -f

echo ""
echo "============================================"
echo "Cleanup complete!"
echo "To remove ALL unused images (even tagged ones like old postgres versions),"
echo "ensure your project is RUNNING and then run:"
echo "  docker image prune -a"
echo "============================================"
