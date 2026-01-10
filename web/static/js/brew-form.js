/**
 * Alpine.js component for the brew form
 * Manages pours, new entity modals, and form state
 * Populates dropdowns from client-side cache for faster UX
 */
function brewForm() {
    return {
        showNewBean: false,
        showNewGrinder: false,
        showNewBrewer: false,
        rating: 5,
        pours: [],
        newBean: { name: '', origin: '', roasterRKey: '', roastLevel: '', process: '', description: '' },
        newGrinder: { name: '', grinderType: '', burrType: '', notes: '' },
        newBrewer: { name: '', description: '' },
        
        // Dropdown data
        beans: [],
        grinders: [],
        brewers: [],
        roasters: [],
        dataLoaded: false,
        
        async init() {
            // Load existing pours if editing
            const poursData = this.$el.getAttribute('data-pours');
            if (poursData) {
                try {
                    this.pours = JSON.parse(poursData);
                } catch (e) {
                    console.error('Failed to parse pours data:', e);
                    this.pours = [];
                }
            }
            
            // Populate dropdowns from cache using stale-while-revalidate pattern
            await this.loadDropdownData();
        },
        
        async loadDropdownData() {
            if (!window.ArabicaCache) {
                console.warn('ArabicaCache not available');
                return;
            }
            
            // First, try to immediately populate from cached data (sync)
            // This prevents flickering by showing data instantly
            const cachedData = window.ArabicaCache.getCachedData();
            if (cachedData) {
                this.applyData(cachedData);
            }
            
            // Then refresh in background if cache is stale
            if (!window.ArabicaCache.isCacheValid()) {
                try {
                    const freshData = await window.ArabicaCache.refreshCache();
                    if (freshData) {
                        this.applyData(freshData);
                    }
                } catch (e) {
                    console.error('Failed to refresh dropdown data:', e);
                    // We already have cached data displayed, so this is non-fatal
                }
            }
        },
        
        applyData(data) {
            this.beans = data.beans || [];
            this.grinders = data.grinders || [];
            this.brewers = data.brewers || [];
            this.roasters = data.roasters || [];
            this.dataLoaded = true;
            
            // Populate the select elements
            this.populateDropdowns();
        },
        
        populateDropdowns() {
            // Get the current selected values (from server-rendered form when editing)
            const beanSelect = this.$el.querySelector('select[name="bean_rkey"]');
            const grinderSelect = this.$el.querySelector('select[name="grinder_rkey"]');
            const brewerSelect = this.$el.querySelector('select[name="brewer_rkey"]');
            
            const selectedBean = beanSelect?.value || '';
            const selectedGrinder = grinderSelect?.value || '';
            const selectedBrewer = brewerSelect?.value || '';
            
            // Populate beans
            if (beanSelect && this.beans.length > 0) {
                // Keep only the first option (placeholder)
                beanSelect.innerHTML = '<option value="">Select a bean...</option>';
                this.beans.forEach(bean => {
                    const option = document.createElement('option');
                    option.value = bean.rkey || bean.RKey;
                    const roasterName = bean.Roaster?.Name || bean.roaster?.name || '';
                    const roasterSuffix = roasterName ? ` - ${roasterName}` : '';
                    option.textContent = `${bean.Name || bean.name} (${bean.Origin || bean.origin} - ${bean.RoastLevel || bean.roast_level})${roasterSuffix}`;
                    option.className = 'truncate';
                    if ((bean.rkey || bean.RKey) === selectedBean) {
                        option.selected = true;
                    }
                    beanSelect.appendChild(option);
                });
            }
            
            // Populate grinders
            if (grinderSelect && this.grinders.length > 0) {
                grinderSelect.innerHTML = '<option value="">Select a grinder...</option>';
                this.grinders.forEach(grinder => {
                    const option = document.createElement('option');
                    option.value = grinder.rkey || grinder.RKey;
                    option.textContent = grinder.Name || grinder.name;
                    option.className = 'truncate';
                    if ((grinder.rkey || grinder.RKey) === selectedGrinder) {
                        option.selected = true;
                    }
                    grinderSelect.appendChild(option);
                });
            }
            
            // Populate brewers
            if (brewerSelect && this.brewers.length > 0) {
                brewerSelect.innerHTML = '<option value="">Select brew method...</option>';
                this.brewers.forEach(brewer => {
                    const option = document.createElement('option');
                    option.value = brewer.rkey || brewer.RKey;
                    option.textContent = brewer.Name || brewer.name;
                    option.className = 'truncate';
                    if ((brewer.rkey || brewer.RKey) === selectedBrewer) {
                        option.selected = true;
                    }
                    brewerSelect.appendChild(option);
                });
            }
            
            // Populate roasters in new bean modal
            const roasterSelect = this.$el.querySelector('select[name="roaster_rkey_modal"]');
            if (roasterSelect && this.roasters.length > 0) {
                roasterSelect.innerHTML = '<option value="">No roaster</option>';
                this.roasters.forEach(roaster => {
                    const option = document.createElement('option');
                    option.value = roaster.rkey || roaster.RKey;
                    option.textContent = roaster.Name || roaster.name;
                    roasterSelect.appendChild(option);
                });
            }
        },
        
        addPour() {
            this.pours.push({ water: '', time: '' });
        },
        
        removePour(index) {
            this.pours.splice(index, 1);
        },
        
        async addBean() {
            if (!this.newBean.name || !this.newBean.origin) {
                alert('Bean name and origin are required');
                return;
            }
            const payload = {
                name: this.newBean.name,
                origin: this.newBean.origin,
                roast_level: this.newBean.roastLevel,
                process: this.newBean.process,
                description: this.newBean.description,
                roaster_rkey: this.newBean.roasterRKey || ''
            };
            const response = await fetch('/api/beans', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (response.ok) {
                const newBean = await response.json();
                // Invalidate cache and refresh data
                if (window.ArabicaCache) {
                    await window.ArabicaCache.invalidateAndRefresh();
                }
                // Reload dropdowns and select the new bean
                await this.loadDropdownData();
                const beanSelect = this.$el.querySelector('select[name="bean_rkey"]');
                if (beanSelect && newBean.rkey) {
                    beanSelect.value = newBean.rkey;
                }
                // Close modal and reset form
                this.showNewBean = false;
                this.newBean = { name: '', origin: '', roasterRKey: '', roastLevel: '', process: '', description: '' };
            } else {
                const errorText = await response.text();
                alert('Failed to add bean: ' + errorText);
            }
        },
        
        async addGrinder() {
            if (!this.newGrinder.name) {
                alert('Grinder name is required');
                return;
            }
            const response = await fetch('/api/grinders', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.newGrinder)
            });
            if (response.ok) {
                const newGrinder = await response.json();
                // Invalidate cache and refresh data
                if (window.ArabicaCache) {
                    await window.ArabicaCache.invalidateAndRefresh();
                }
                // Reload dropdowns and select the new grinder
                await this.loadDropdownData();
                const grinderSelect = this.$el.querySelector('select[name="grinder_rkey"]');
                if (grinderSelect && newGrinder.rkey) {
                    grinderSelect.value = newGrinder.rkey;
                }
                // Close modal and reset form
                this.showNewGrinder = false;
                this.newGrinder = { name: '', grinderType: '', burrType: '', notes: '' };
            } else {
                const errorText = await response.text();
                alert('Failed to add grinder: ' + errorText);
            }
        },
        
        async addBrewer() {
            if (!this.newBrewer.name) {
                alert('Brewer name is required');
                return;
            }
            const response = await fetch('/api/brewers', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.newBrewer)
            });
            if (response.ok) {
                const newBrewer = await response.json();
                // Invalidate cache and refresh data
                if (window.ArabicaCache) {
                    await window.ArabicaCache.invalidateAndRefresh();
                }
                // Reload dropdowns and select the new brewer
                await this.loadDropdownData();
                const brewerSelect = this.$el.querySelector('select[name="brewer_rkey"]');
                if (brewerSelect && newBrewer.rkey) {
                    brewerSelect.value = newBrewer.rkey;
                }
                // Close modal and reset form
                this.showNewBrewer = false;
                this.newBrewer = { name: '', description: '' };
            } else {
                const errorText = await response.text();
                alert('Failed to add brewer: ' + errorText);
            }
        }
    }
}
